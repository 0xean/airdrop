
alter table chains add column decimals integer;
---- create above / drop below ----
-- undo --
alter table chains drop column decimals;
