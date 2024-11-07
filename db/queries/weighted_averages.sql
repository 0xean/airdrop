		with params as (
			select
				chains.name,
				tokens.address as token_address,
				tokens.min_eligible,
				snapshot_start_block,
				snapshot_end_block
			from
				chains
				join tokens on chains.name = tokens.chain
			where
				chains.name = 'ETH'
				and tokens.symbol = 'FOX'
		),
		holders as (
			select
				distinct transfer_to as account
			from
				transfers
			where
				token = (
					select
						token_address
					from
						params
				)
		),
		token_transfers as (
			select
				id,
				txhash,
				transfer_to as account,
				transfer_value as delta,
				block_number
			from
				holders
				join transfers on holders.account = transfers.transfer_to
			where
				token = (
					select
						token_address
					from
						params
				)
				and transfer_from != transfer_to
			union
			select
				id,
				txhash,
				transfer_from,
				-(transfer_value),
				block_number
			from
				holders
				join transfers on holders.account = transfers.transfer_from
			where
				token = (
					select
						token_address
					from
						params
				)
				and transfer_from != transfer_to
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
								delta,
								block_number
							from
								token_transfers
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
							union
							-- starting balance
							(
								select
									account,
									sum(delta),
									(
										select
											snapshot_start_block
										from
											params
									) as block_number
								from
									token_transfers
								where
									block_number <= (
										select
											snapshot_start_block
										from
											params
									)
								group by
									account
								order by
									block_number
							)
							union
							-- ending balance
							(
								select
										distinct account,
										0,
										(
												select
														snapshot_end_block
												from
														params
										) as block_number
								from
										token_transfers
								where
										block_number <= (
												select
														snapshot_end_block
												from
														params
										)
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
		where account = '0xaf2466d9492d6372d57c6ec666495690133538f4'
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
		) > (select min_eligible from params)
		order by
			avg_hold desc;
