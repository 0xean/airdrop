package indexer

import (
	"fmt"

	"github.com/ArkeoNetwork/airdrop/pkg/db"
	"github.com/ArkeoNetwork/common/logging"
	"github.com/ArkeoNetwork/common/utils"
)

type IndexerAppParams struct {
	SnapshotStart uint64
	SnapshotEnd   uint64
	utils.DBConfig
}

type IndexerApp struct {
	params IndexerAppParams
	db     *db.AirdropDB
}

var log = logging.WithoutFields()

func NewIndexer(params IndexerAppParams) *IndexerApp {
	d, err := db.New(params.DBConfig)
	if err != nil {
		panic(fmt.Sprintf("error connecting to the db: %+v", err))
	}
	return &IndexerApp{params: params, db: d}
}

// index transfers, staking rewards, hedgeys
func (app *IndexerApp) Start() {
	log.Info("starting indexing transfers")

	err := app.IndexTransfers()
	if err != nil {
		panic(fmt.Sprintf("error indexing transfers: %+v", err))
	}
	log.Info("finished indexing transfers")
	log.Info("starting indexing LP staking")

	err = app.IndexStakingRewardsEvents("stakingrewardsv5")
	if err != nil {
		panic(fmt.Sprintf("error indexing LP staking v5: %+v", err))
	}
	err = app.IndexStakingRewardsEvents("stakingrewardsv6")
	if err != nil {
		panic(fmt.Sprintf("error indexing LP staking v6: %+v", err))
	}
	err = app.IndexStakingRewardsEvents("stakingrewardsv7")
	if err != nil {
		panic(fmt.Sprintf("error indexing LP staking v7: %+v", err))
	}
	err = app.IndexStakingRewardsEvents("stakingrewardsv8")
	if err != nil {
		panic(fmt.Sprintf("error indexing LP staking v8: %+v", err))
	}
	err = app.IndexStakingRewardsEvents("stakingrewardsv9")
	if err != nil {
		panic(fmt.Sprintf("error indexing LP staking v9: %+v", err))
	}
	log.Info("finished indexing LP staking")
	log.Info("starting indexing hedgeys")

	err = app.IndexHedgeyEvents()
	if err != nil {
		panic(fmt.Sprintf("error indexing hedgeys: %+v", err))
	}
}
