package db

const (
	sqlFindLatestThorchainLPBlockIndexed = `
		select coalesce(max(block_number),0) from thorchain_lp_events where chain = $1 and pool = $2
	`
	sqlInsertThorLPBalanceEvent = `
		insert into thorchain_lp_events(chain,block_number,pool,balance_asset,balance_rune,address_thor,address_asset)
		values ($1,$2,$3,$4,$5,$6,$7)
		on conflict on constraint thorchain_lp_events_unique
			do nothing
	`
	sqlAverageThorchainLP = `
		with unique_providers as (
			select pool, address_asset as account, min(block_number),max(block_number), count(1) as sample_count
			from thorchain_lp_events
			where address_thor = ''
			group by pool, address_asset
			union all
			select pool, address_thor, min(block_number),max(block_number), count(1)
			from thorchain_lp_events
			where address_asset = ''
			group by pool, address_thor
			union all
			select pool, address_asset, min(block_number),max(block_number), count(1)
			from thorchain_lp_events
			where address_asset != ''
				and address_thor != ''
			group by pool, address_asset
			),
			sample_count as (
			select count(distinct block_number) total
			from thorchain_lp_events
			)
			select p.account, sum(balance_asset) / (select sample_count.total from sample_count) as average_lp_asset
			from unique_providers p join thorchain_lp_events evts on p.pool = evts.pool and p.account = evts.address_asset
			-- where p.pool = 'ETH.FOX-0XC770EEFAD204B5180DF6A14EE197D99D808EE52D'
			where p.pool = $1
			group by p.account
			order by average_lp_asset desc
	`
)
