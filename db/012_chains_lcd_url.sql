alter table chains add column lcd_url text not null default '';
update chains set lcd_url = 'https://cosmos-rest.publicnode.com' where name = 'GAIA';
update chains set lcd_url = 'https://daemon.thorchain.shapeshift.com:443/lcd' where name = 'THOR';
---- create above / drop below ----
alter table chains drop column lcd_url;
