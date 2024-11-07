package db

const (
	sqlFindCosmosStakingAveragedBalances = `
		with params as (
			select
				chains.name,
				snapshot_start_block,
				snapshot_end_block,
				min_eligible
			from
				chains
			where
				chains.name = $1
		),
		staking_events as (
			select
				id,
				txhash,
				delegator as account,
				validator,
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
				account,
				cumulative_balance,
				block_number,
				blocks_between,
				blocks_between * lag(cumulative_balance, 1, 0) over (partition by account) as avg_over_blocks
			from
				(
					select
						account,
						SUM(delta) over (
							partition by account
							order by
								block_number
						) as cumulative_balance,
						ts.block_number,
						block_number - lag(block_number, 1, block_number) over (
							partition by account
							order by
								block_number
						) as blocks_between
					from
						(
							select
								account,
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
								and block_number < (
									select
										snapshot_end_block
									from
										params
								)
							group by
								account,
								block_number
							union
							-- dummy zero delta row at end height so calculate proper blocks between for last record
							select
								account,
								0 as delta,
								(
									select
										snapshot_end_block
									from
										params
								) as block_number
							from
								staking_events
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
			account,
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
			account
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
					min_eligible
				from
					params
			)
		order by
		avg_hold desc
	`
)
