with params as (
	select
		chains.name,
		snapshot_start_block,
		snapshot_end_block,
		100 as min_balance
	from
		chains
	where
		chains.name = 'GAIA'
),
staking_events as (
	select
		id,
		txhash,
		validator,
		delegator as account,
		amount as delta,
		block_number
	from
		cosmos_staking_events
	where
		chain = (
			select
				chain
			from
				params
		)
	order by
		block_number
),
averageable as (
	select
		validator,
		cumulative_balance,
		block_number,
		blocks_between,
		blocks_between * lag(cumulative_balance, 1, 0) over (partition by validator) as avg_over_blocks
	from
		(
			select
				validator,
				SUM(delta) over (
					partition by validator
					order by
						block_number
				) as cumulative_balance,
				ts.block_number,
				block_number - lag(block_number, 1, block_number) over (
					partition by validator
					order by
						block_number
				) as blocks_between
			from
				(
					select
						validator,
						sum(delta) as delta,
						block_number
					from
						staking_events
					where
						block_number >= (
							select
								snapshot_start_block
							from
								params
						)
						and block_number <= (
							select
								snapshot_end_block
							from
								params
						)
					group by
						validator,
						block_number
					union
					-- starting balance
					(
						select
							validator,
							sum(delta),
							(
								select
									snapshot_start_block
								from
									params
							) as block_number
						from
							staking_events
						where
							block_number <= (
								select
									snapshot_start_block
								from
									params
							)
						group by
							validator
						order by
							block_number
					)
					union
					-- ending balance
					(
						select
							validator,
							sum(staking_events.delta),
							(
								select
									snapshot_end_block
								from
									params
							) as block_number
						from
							staking_events
						where
							block_number >= (
								select
									snapshot_start_block
								from
									params
							)
							and block_number <= (
								select
									snapshot_end_block
								from
									params
							)
						group by
							validator
						order by
							block_number
					)
				) as ts
			where
				ts.block_number >= (
					select
						snapshot_start_block
					from
						params
				)
				and ts.block_number <= (
					select
						snapshot_end_block
					from
						params
				)
		) as x
	order by
		block_number
)
select
	validator,
	sum(avg_over_blocks) / (
		(
			select
				snapshot_end_block
			from
				params
		) - (
			select
				snapshot_start_block
			from
				params
		)
	) avg_hold
from
	averageable
group by
	validator
having
	sum(avg_over_blocks) / (
		(
			select
				snapshot_end_block
			from
				params
		) - (
			select
				snapshot_start_block
			from
				params
		)
	) > (
		select
			min_balance
		from
			params
	)
order by
	avg_hold desc;