package db

var (
	sqlFindAllChains = `select name,rpc_url,lcd_url,snapshot_start_block,snapshot_end_block,coalesce(decimals,0) as decimals from chains`
	sqlFindChain     = `select name,rpc_url,lcd_url,snapshot_start_block,snapshot_end_block,coalesce(decimals,0) as decimals from chains where name = $1`
	sqlFindEthChains = `select name,rpc_url,lcd_url,snapshot_start_block,snapshot_end_block,coalesce(decimals,0) as decimals from chains where name in ('ETH','POLY','GNO')` // ('ETH','POLY','GNO')
)
