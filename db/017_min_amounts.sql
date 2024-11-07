alter table chains add column min_eligible numeric not null default 0;
alter table tokens add column min_eligible numeric not null default 0;

---- create above / drop below ----
alter table chains drop column min_eligible;
alter table tokens drop column min_eligible;
