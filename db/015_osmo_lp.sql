create table osmo_lp (
  id bigserial not null constraint osmo_lp_pk primary key,
  created timestamptz default now() not null,
  updated timestamptz default now() not null,
  block_number numeric not null check (block_number >= 0),
  account text not null,
  qty_osmo numeric not null,
  constraint osmo_lp_unique unique (
    block_number,
    account
  )
);

---- create above / drop below ----
drop table osmo_lp;
