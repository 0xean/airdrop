package indexer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	"github.com/pkg/errors"
)

type ThorPool struct {
	Asset               string `json:"asset"`
	BalanceAsset        int64  `json:"balance_asset,string"`
	BalanceRune         int64  `json:"balance_rune,string"`
	PoolUnits           int64  `json:"pool_units,string"`
	LpUnits             int64  `json:"LP_units,string"`
	Status              string `json:"status"`
	SynthSupply         int64  `json:"synth_supply,string"`
	SynthUnits          int64  `json:"synth_units,string"`
	PendingInboundRune  int64  `json:"pending_inbound_rune,string"`
	PendingInboundAsset int64  `json:"pending_inbound_asset,string"`
}

type ThorLiquidityProvider struct {
	Asset             string `json:"asset"`
	AddressThor       string `json:"rune_address"`
	AddressNative     string `json:"asset_address"`
	LastAddHeight     int64  `json:"last_add_height"`
	Units             int64  `json:"units,string"`
	PendingRune       int64  `json:"pending_rune,string"`
	PendingAsset      int64  `json:"pending_asset,string"`
	RuneDepositValue  int64  `json:"rune_deposit_value,string"`
	AssetDepositValue int64  `json:"asset_deposit_value,string"`
}

func (c *CosmosIndexer) IndexThorLP(poolName string) error {
	startHeight := int64(c.chain.SnapshotStartBlock)
	endHeight := int64(c.chain.SnapshotEndBlock)

	latestBlock, err := c.tm.Block(context.Background(), nil)
	if err != nil {
		return errors.Wrapf(err, "error getting latest block")
	}

	if endHeight > latestBlock.Block.Height {
		endHeight = latestBlock.Block.Height
	}

	latest, err := c.db.FindLatestIndexedThorchainLPBlock(c.chain.Name, poolName)
	if err != nil {
		return errors.Wrapf(err, "error finding latest indexed block")
	}
	if latest > startHeight {
		log.Infof("found latest indexed block %d, starting at %d", latest, latest-1)
		startHeight = latest - 1
	}

	blocksPerDay := int64(6 * 60 * 24)
	// get up to an extra day of blocks, trim if post cutoff
	for i := startHeight; i <= endHeight+blocksPerDay; i += blocksPerDay {
		if _, err := c.indexThorLP(i, poolName); err != nil {
			if err.Error() == "no pool" {
				break
			}
			log.Errorf("error indexing delegations at height %d: %+v", i, err)
			break
		}
		log.Infof("indexed %s lp block %d", poolName, i)
	}

	return nil
}

func (c *CosmosIndexer) indexThorLP(height int64, poolName string) ([]types.ThorLPBalanceEvent, error) {
	pool, err := c.findThorPoolByHeight(height, poolName)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading pool at height %d for %s", height, poolName)
	}

	if pool == nil {
		log.Infof("pool %s not found at height %d", poolName, height)
		return nil, fmt.Errorf("no pool")
	}

	totalUnits := pool.PoolUnits
	lpBalances, err := c.findThorLpsByHeight(height, poolName)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading balances at height %d for %s", height, poolName)
	}

	batch := make([]types.ThorLPBalanceEvent, 0, len(lpBalances))
	for _, bal := range lpBalances {
		share := float64(bal.Units) / float64(totalUnits)
		shareAsset := share * utils.FromBaseUnits(pool.BalanceAsset, 8)
		shareRune := share * utils.FromBaseUnits(pool.BalanceRune, 8)
		batch = append(batch,
			types.ThorLPBalanceEvent{
				Chain:         c.chain.Name,
				BlockNumber:   int64(height),
				Pool:          poolName,
				AddressThor:   bal.AddressThor,
				AddressNative: bal.AddressNative,
				BalanceRune:   shareRune,
				BalanceAsset:  shareAsset,
			},
		)
	}
	if err = c.db.InsertThorLPBalanceEvent(batch); err != nil {
		return nil, errors.Wrapf(err, "error inserting thor lp balances at height %d for %s", height, poolName)
	}
	return batch, nil
}

// find lp balances at a given height
// Savers almost identical: /thorchain/pool/{pool}/savers
func (c *CosmosIndexer) findThorLpsByHeight(height int64, poolName string) ([]ThorLiquidityProvider, error) {
	path := fmt.Sprintf("/thorchain/pool/%s/liquidity_providers", poolName)
	res, err := c.lcd.R().SetQueryParam("height", fmt.Sprintf("%d", height)).Get(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading balances at height %d, pool %s", height, poolName)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("error reading balances at height %d, pool %s, status code %d", height, poolName, res.StatusCode())
	}

	results := make([]ThorLiquidityProvider, 0, 1024)
	if err = json.Unmarshal(res.Body(), &results); err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling balances at height %d, pool %s", height, poolName)
	}
	return results, nil
}

func (c *CosmosIndexer) findThorPoolByHeight(height int64, poolName string) (*ThorPool, error) {
	pool := ThorPool{}
	path := fmt.Sprintf("/thorchain/pool/%s", poolName)
	res, err := c.lcd.R().SetQueryParam("height", fmt.Sprintf("%d", height)).Get(path)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading pool at height %d, pool %s", height, poolName)
	}
	switch res.StatusCode() {
	case 200:
	case 404:
		return nil, nil
	default:
		return nil, fmt.Errorf("error reading pool at height %d, pool %s, status code %d", height, poolName, res.StatusCode())
	}

	if err = json.Unmarshal(res.Body(), &pool); err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling pool at height %d, pool %s", height, poolName)
	}

	return &pool, nil
}
