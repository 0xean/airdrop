package db

const (
	sqlInsertCosmosStakingEvent = `
		insert into cosmos_staking_events(chain, event_type, delegator, validator, amount, block_number, txhash, event_index)
		values ($1,$2,$3,$4,$5,$6,$7,$8)
		on conflict on constraint cosmos_staking_events_unique
		do nothing
	`
	sqlFindLatestCosmosStakingBlockIndexed = `
		select coalesce(max(block_number),0) from cosmos_staking_events where chain = $1
	`
)
