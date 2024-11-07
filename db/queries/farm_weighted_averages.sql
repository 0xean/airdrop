with params as (
    select
        chains.name,
        contracts.address as contract_address,
        contract_name as contract_name,
        contracts.genesis_block,
        tokens.address as token_address,
        tokens.min_eligible,
        snapshot_start_block,
        snapshot_end_block
    from
        chains
        join staking_contracts as contracts on chains.name = contracts.chain
        join tokens on chains.name = tokens.chain
    where
        chains.name = 'ETH' -- $1
        and contracts.contract_name = 'stakingrewards' -- $2
        and tokens.symbol = 'UNI-V2' -- $3
),
filtered_staking_events as (
    select
        *
    from
        staking_events evts
    where
        evts.block_number >= (
            select
                genesis_block
            from
                params
        )
        and evts.chain = (
            select
                name
            from
                params
        )
        and evts.staking_contract = (
            select
                contract_address
            from
                params
        )
        and evts.token = (
            select
                token_address
            from
                params
        ) -- 			    and evts.staker = '0x2d854fbda34f3a1e2b55a8b911e786ba5afdc3e0'
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
                    -- apply staking events
                    select
                        staker as account,
                        stake_value as delta,
                        block_number
                    from
                        filtered_staking_events
                    where
                        block_number > (
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
                            staker,
                            sum(stake_value),
                            (
                                select
                                    snapshot_start_block
                                from
                                    params
                            ) as block_number
                        from
                            filtered_staking_events
                        where
                            block_number <= (
                                select
                                    snapshot_start_block
                                from
                                    params
                            )
                        group by
                            staker
                        order by
                            block_number
                    )
                    union
                    -- ending balance
                    (
                        select
                            distinct staker,
                            0,
                            (
                                select
                                    snapshot_end_block
                                from
                                    params
                            ) as block_number
                        from
                            filtered_staking_events
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
    avg_hold desc;