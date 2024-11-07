create table cosmos_staking_events
(
    id           bigserial not null
                    constraint cosmos_staking_events_pk
                        primary key,
    created      timestamptz default now() not null,
    updated      timestamptz default now() not null,
    chain        text not null,
    event_type   text not null,
    delegator    text not null,
    validator    text not null,
    amount       numeric not null,
    block_number numeric not null check ( block_number >= 0 ),
    txhash       text not null,
    event_index  integer not null,
    constraint cosmos_staking_events_unique
        unique (chain,txhash,event_index,validator,delegator)
);

---- create above / drop below ----
-- undo --
drop table cosmos_staking_events;
