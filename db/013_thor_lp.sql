create table thorchain_lp_events (
    id bigserial not null constraint thorchain_lp_events_pk primary key,
    created timestamptz default now() not null,
    updated timestamptz default now() not null,
    chain text not null,
    block_number numeric not null check (block_number >= 0),
    pool text not null,
    balance_asset numeric not null, 
    balance_rune numeric not null,
    address_asset text,
    address_thor text,
    constraint thorchain_lp_events_unique unique (chain, block_number, pool, address_thor,address_asset)
);

---- create above / drop below ----
-- undo --
drop table thorchain_lp_events;
