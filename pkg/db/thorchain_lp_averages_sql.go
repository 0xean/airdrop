package db

const (
	sqlFindAveragedThorLPBalances = `
		with params as (
			select
				$1 as pool
		),
		unique_providers as (
			select
				pool,
				address_asset as account,
				min(block_number),
				max(block_number),
				count(1) as sample_count
			from
				thorchain_lp_events
			where
				address_thor = ''
			group by
				pool,
				address_asset
			union
			all
			select
				pool,
				address_thor,
				min(block_number),
				max(block_number),
				count(1)
			from
				thorchain_lp_events
			where
				address_asset = ''
			group by
				pool,
				address_thor
			union
			all
			select
				pool,
				address_asset,
				min(block_number),
				max(block_number),
				count(1)
			from
				thorchain_lp_events
			where
				address_asset != ''
				and address_thor != ''
			group by
				pool,
				address_asset
		),
		sample_count as (
			select
				count(distinct block_number) total
			from
				thorchain_lp_events
			where
				pool = (select pool from params)
		)
		select
			p.account,
			sum(balance_asset) / (
				select
					sample_count.total
				from
					sample_count
			) as avg_hold
		from
			unique_providers p
			join thorchain_lp_events evts on p.pool = evts.pool
			and p.account = evts.address_asset
		where
			p.pool = (select pool from params)
		group by
			p.account
		order by
		avg_hold desc
	`
)
