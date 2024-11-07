ALTER TABLE osmo_lp RENAME COLUMN qty_osmo TO lp_amount;

ALTER TABLE osmo_lp ADD COLUMN pool_id numeric;
