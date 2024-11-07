package db

const (
	sqlFindAveragedBalances = `
	WITH params AS (
		SELECT
			chains.name,
			tokens.address AS token_address,
			tokens.min_eligible,
			snapshot_start_block,
			snapshot_end_block
		FROM
			chains
			JOIN tokens ON chains.name = tokens.chain
		WHERE
			chains.name = $1
			AND tokens.symbol = $2
	),
	holders AS (
		SELECT
			DISTINCT transfer_to AS account
		FROM
			transfers
		WHERE
			token = (
				SELECT
					token_address
				FROM
					params
			)
	),
	token_transfers AS (
		SELECT
			id,
			txhash,
			transfer_to AS account,
			transfer_value AS delta,
			block_number
		FROM
			holders
			JOIN transfers ON holders.account = transfers.transfer_to
		WHERE
			token = (
				SELECT
					token_address
				FROM
					params
			)
			AND transfer_from != transfer_to
		UNION
		SELECT
			id,
			txhash,
			transfer_from,
			-(transfer_value),
			block_number
		FROM
			holders
			JOIN transfers ON holders.account = transfers.transfer_from
		WHERE
			token = (
				SELECT
					token_address
				FROM
					params
			)
			AND transfer_from != transfer_to
		ORDER BY
			block_number
	),
	averageable AS (
		SELECT
			account,
			cumulative_balance,
			block_number,
			blocks_between,
			blocks_between * LAG(cumulative_balance, 1, 0) OVER (PARTITION BY account) AS avg_over_blocks
		FROM
			(
				SELECT
					account,
					SUM(delta) OVER (
						PARTITION BY account
						ORDER BY
							block_number
					) AS cumulative_balance,
					ts.block_number,
					block_number - LAG(block_number, 1, block_number) OVER (
						PARTITION BY account
						ORDER BY
							block_number
					) AS blocks_between
				FROM
					(
						SELECT
							account,
							delta,
							block_number
						FROM
							token_transfers
						WHERE
							block_number >= (
								SELECT
									snapshot_start_block
								FROM
									params
							)
							AND block_number <= (
								SELECT
									snapshot_end_block
								FROM
									params
							)
						UNION
						-- starting balance
						(
							SELECT
								account,
								SUM(delta),
								(
									SELECT
										snapshot_start_block
									FROM
										params
								) AS block_number
							FROM
								token_transfers
							WHERE
								block_number <= (
									SELECT
										snapshot_start_block
									FROM
										params
								)
							GROUP BY
								account
							ORDER BY
								block_number
						)
						UNION
						-- ending balance
						(
							SELECT
								DISTINCT account,
								0,
								(
									SELECT
										snapshot_end_block
									FROM
										params
								) AS block_number
							FROM
								token_transfers
							WHERE
								block_number <= (
									SELECT
										snapshot_end_block
									FROM
										params
								)
							ORDER BY
								block_number
						)
					) AS ts
				WHERE
					ts.block_number >= (
						SELECT
							snapshot_start_block
						FROM
							params
					)
					AND ts.block_number <= (
						SELECT
							snapshot_end_block
						FROM
							params
					)
			) AS x
		ORDER BY
			block_number
	)
	SELECT
		account,
		CASE
			WHEN sv.address IS NOT NULL THEN 2 * SUM(avg_over_blocks) / (
				(
					SELECT
						snapshot_end_block
					FROM
						params
				) - (
					SELECT
						snapshot_start_block
					FROM
						params
				)
			)
			ELSE SUM(avg_over_blocks) / (
				(
					SELECT
						snapshot_end_block
					FROM
						params
				) - (
					SELECT
						snapshot_start_block
					FROM
						params
				)
			)
		END AS avg_hold
	FROM
		averageable a
		LEFT JOIN snapshot_voters sv ON a.account = sv.address
	GROUP BY
		account, sv.address
	HAVING
		SUM(avg_over_blocks) / (
			(
				SELECT
					snapshot_end_block
				FROM
					params
			) - (
				SELECT
					snapshot_start_block
				FROM
					params
			)
		) > (
			SELECT
				min_eligible
			FROM
				params
		)
	ORDER BY
		avg_hold DESC	
	`
	// average staked/farmed balances for eth
	sqlFindAveragedFarmBalances = `
	WITH params AS (
		SELECT
			chains.name,
			contracts.address AS contract_address,
			contract_name AS contract_name,
			contracts.genesis_block,
			tokens.address AS token_address,
			tokens.min_eligible,
			snapshot_start_block,
			snapshot_end_block
		FROM
			chains
			JOIN staking_contracts AS contracts ON chains.name = contracts.chain
			JOIN tokens ON chains.name = tokens.chain
		WHERE
			chains.name = $1 -- 'ETH'
			AND contracts.contract_name = $2 -- 'stakingrewards'
			AND tokens.symbol = $3 -- 'UNI-V2'
	),
	filtered_staking_events AS (
		SELECT
			*
		FROM
			staking_events evts
		WHERE
			evts.block_number >= (
				SELECT
					genesis_block
				FROM
					params
			)
			AND evts.chain = (
				SELECT
					name
				FROM
					params
			)
			AND evts.staking_contract = (
				SELECT
					contract_address
				FROM
					params
			)
			AND evts.token = (
				SELECT
					token_address
				FROM
					params
			)
	),
	averageable AS (
		SELECT
			account,
			cumulative_balance,
			block_number,
			blocks_between,
			blocks_between * LAG(cumulative_balance, 1, 0) OVER (PARTITION BY account) AS avg_over_blocks
		FROM
			(
				SELECT
					account,
					SUM(delta) OVER (
						PARTITION BY account
						ORDER BY
							block_number
					) AS cumulative_balance,
					ts.block_number,
					block_number - LAG(block_number, 1, block_number) OVER (
						PARTITION BY account
						ORDER BY
							block_number
					) AS blocks_between
				FROM
					(
						-- apply staking events
						SELECT
							staker AS account,
							stake_value AS delta,
							block_number
						FROM
							filtered_staking_events
						WHERE
							block_number > (
								SELECT
									snapshot_start_block
								FROM
									params
							)
							AND block_number <= (
								SELECT
									snapshot_end_block
								FROM
									params
							)
						UNION
						-- starting balance
						(
							SELECT
								staker,
								SUM(stake_value),
								(
									SELECT
										snapshot_start_block
									FROM
										params
								) AS block_number
							FROM
								filtered_staking_events
							WHERE
								block_number <= (
									SELECT
										snapshot_start_block
									FROM
										params
								)
							GROUP BY
								staker
							ORDER BY
								block_number
						)
						UNION
						-- ending balance
						(
							SELECT
								DISTINCT staker,
								0,
								(
									SELECT
										snapshot_end_block
									FROM
										params
								) AS block_number
							FROM
								filtered_staking_events
							WHERE
								block_number <= (
									SELECT
										snapshot_end_block
									FROM
										params
								)
							ORDER BY
								block_number
						)
					) AS ts
				WHERE
					ts.block_number >= (
						SELECT
							snapshot_start_block
						FROM
							params
					)
					AND ts.block_number <= (
						SELECT
							snapshot_end_block
						FROM
							params
					)
			) AS x
		ORDER BY
			block_number
	)
	SELECT
		account,
		CASE
			WHEN sv.address IS NOT NULL THEN 2 * sum(avg_over_blocks) / (
				(
					SELECT
						snapshot_end_block
					FROM
						params
				) - (
					SELECT
						snapshot_start_block
					FROM
						params
				)
			)
			ELSE sum(avg_over_blocks) / (
				(
					SELECT
						snapshot_end_block
					FROM
						params
				) - (
					SELECT
						snapshot_start_block
					FROM
						params
				)
			)
		END AS avg_hold
	FROM
		averageable a
		LEFT JOIN snapshot_voters sv ON a.account = sv.address
	GROUP BY
		account, sv.address
	HAVING
		sum(avg_over_blocks) / (
			(
				SELECT
					snapshot_end_block
				FROM
					params
			) - (
				SELECT
					snapshot_start_block
				FROM
					params
			)
		) > (
			SELECT
				min_eligible
			FROM
				params
		)
	ORDER BY
		avg_hold DESC
	`
)
