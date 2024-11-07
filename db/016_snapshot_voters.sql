create table snapshot_voters
(
    id      bigserial not null
            constraint snapshot_voters_pk
            primary key,
    created timestamptz default now() not null,
    updated timestamptz default now() not null,
    address text not null unique,
    constraint snapshot_voters_address_unique
        unique (address)
);

---- create above / drop below ----
-- undo --
drop table snapshot_voters;