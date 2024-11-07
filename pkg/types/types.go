package types

type Token struct {
	Address      string `db:"address"`
	Name         string `db:"name"`
	Symbol       string `db:"symbol"`
	Decimals     uint64 `db:"decimals"`
	Chain        string `db:"chain"`
	GenesisBlock uint64 `db:"genesis_block"`
	Height       uint64 `db:"height"`
}

type Chain struct {
	Name               string `db:"name"`
	RpcUrl             string `db:"rpc_url"`
	SnapshotStartBlock uint64 `db:"snapshot_start_block"`
	SnapshotEndBlock   uint64 `db:"snapshot_end_block"`
	Decimals           uint8  `db:"decimals"`
	LcdUrl             string `db:"lcd_url"` // cosmos chains
}

type Transfer struct {
	TxHash       string  `db:"txHash"`
	LogIndex     uint64  `db:"log_index"`
	TokenAddress string  `db:"token"`
	From         string  `db:"transfer_from"`
	To           string  `db:"transfer_to"`
	Value        float64 `db:"transfer_value"` // decimal version of transfer_value
	BlockNumber  uint64  `db:"block_number"`
}

type StakingContract struct {
	Address      string `db:"address"`
	ContractName string `db:"contract_name"`
	Chain        string `db:"chain"`
	GenesisBlock uint64 `db:"genesis_block"`
	Height       uint64 `db:"height"`
}

type StakingEvent struct {
	TxHash          string  `db:"txhash"`
	LogIndex        uint    `db:"log_index"`
	Token           string  `db:"token"`
	StakingContract string  `db:"staking_contract"`
	Staker          string  `db:"staker"`
	Value           float64 `db:"stake_value"` // decimal version of value, can be negative for unstake
	BlockNumber     uint64  `db:"block_number"`
	Chain           string  `db:"chain"`
}

type SnapshotVoter struct {
	Address string `db:"address"`
}

type CosmosStakingEvent struct {
	EventType   string  `db:"event_type"`
	Chain       string  `db:"chain"`
	Delegator   string  `db:"delegator"`
	Validator   string  `db:"validator"`
	Value       float64 `db:"amount"`
	BlockNumber uint64  `db:"block_number"`
	TxHash      string  `db:"txhash"`
	EventIndex  int64   `db:"event_index"`
}

type ThorLPBalanceEvent struct {
	Chain         string  `db:"chain"`
	BlockNumber   int64   `db:"block_number"`
	Pool          string  `db:"pool"`
	AddressThor   string  `db:"address_thor"`
	AddressNative string  `db:"address_native"`
	BalanceRune   float64 `db:"balance_rune"`
	BalanceAsset  float64 `db:"balance_asset"`
}

type OsmoLP struct {
	BlockNumber int64  `db:"block_number"`
	Account     string `db:"account"`
	LpAmount    string `db:"lp_amount"`
	PoolId      int64  `db:"pool_id"`
	Type        string `db:"lp_type"`
	TxHash      string `db:"tx_hash"`
}
