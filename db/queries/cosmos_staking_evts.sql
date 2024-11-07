with params as (
  select
    chains.name,
    'osmovaloper1xf9zpq5kpxks49cg606tzd8qstaykxgt2vs0d5' as validator,
    snapshot_start_block,
    snapshot_end_block
  from
    chains
  where
    chains.name = 'OSMO'
)
select
    id,
    txhash,
    delegator as account,
    amount as delta,
    block_number
  from
     cosmos_staking_events
  where
    chain = (select chain from params)
    and validator = (select validator from params)
    and event_type != 'initial'
  order by
    block_number;