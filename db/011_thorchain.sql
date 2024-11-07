insert into chains(name,rpc_url,snapshot_start_block,snapshot_end_block,decimals)
values ('THOR','https://daemon.thorchain.shapeshift.com:443/rpc',8336490,15475617,8);

---- create above / drop below ----
delete from chains where name = 'THOR';
