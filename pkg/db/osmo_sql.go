package db

const (
	sqlInsertOsmoLP = `
		insert into osmo_lp(block_number, account, lp_amount, pool_id, tx_hash, lp_type)
		values ($1,$2,$3,$4,$5,$6)
	`
)
