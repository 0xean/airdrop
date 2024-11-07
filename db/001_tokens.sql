create table tokens
(
     id  bigserial not null
        constraint tokens_pk
            primary key,
    created               timestamptz default now() not null,
    updated               timestamptz default now() not null,
    address text  not null unique,
    name     text   not null,
    symbol   text   not null,
    decimals numeric check ( decimals >= 0 ),
    chain    text   not null,
    genesis_block numeric check ( genesis_block >= 0 ),
    height    numeric not null check ( height >= 0 )

);

-- mainnet 
  -- fox
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0xc770eefad204b5180df6a14ee197d99d808ee52d', 'FOX', 'FOX', 18, 'ETH', 7446830, 0);
  -- foxy
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0xdc49108ce5c57bc3408c3a5e95f3d864ec386ed3', 'FOX Yieldy', 'FOXy', 18, 'ETH', 14476930, 0);
  -- fox <> ETH LP
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0x470e8de2ebaef52014a47cb5e6af86884947f08c', 'Uniswap V2', 'UNI-V2', 18, 'ETH', 10104463, 0);
  -- tfox
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0x808d3e6b23516967ceae4f17a5f9038383ed5311', 'TokemaktFOX', 'tFOX', 18, 'ETH', 13690577, 0);
  -- scfox
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0x04979cccc2f854e167daad0ae095da49eae4842e', 'FOX $0.80 Success Token March 2024', 'scFOX0324', 18, 'ETH', 14174869, 0);

-- gnosis
  -- fox
insert into tokens(address, name, symbol,decimals,  chain, genesis_block, height) values ('0x21a42669643f45bc0e086b8fc2ed70c23d67509d', 'FOX on GNO', 'FOX', 18, 'GNO', 16988521, 0);

-- Polygon
  -- fox
insert into tokens(address, name, symbol, decimals, chain, genesis_block, height) values ('0x65a05db8322701724c197af82c9cae41195b0aa8', 'FOX', 'FOX', 18, 'POLY', 14283484 , 0);

create table staking
(
     id  bigserial not null
        constraint staking_pk
            primary key,
    created               timestamptz default now() not null,
    updated               timestamptz default now() not null,
    address text  not null,
    contract_name   text   not null,
    chain    text   not null,
    genesis_block numeric check ( genesis_block >= 0 ),
    height    numeric not null check ( height >= 0 ),
    constraint address_chain_unique
        unique (address, chain)
);

-- mainnet 
  -- LP stakingrewards
insert into staking(address, contract_name, chain, genesis_block, height) values ('0xc14eaa8284feff79edc118e06cadbf3813a7e555', 'stakingrewardsv5', 'ETH', 15941059, 0);

---- create above / drop below ----
-- undo --
drop table tokens;
drop table staking;

