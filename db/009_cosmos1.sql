insert into chains(name,rpc_url,snapshot_start_block,snapshot_end_block,decimals)
values ('GAIA','https://daemon.cosmos.shapeshift.com:443/rpc',12939961,19931432,6);

---- create above / drop below ----
delete from chains where name = 'GAIA';
