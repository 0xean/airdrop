package snapshot

import (
	"fmt"

	"github.com/ArkeoNetwork/airdrop/pkg/db"
	"github.com/ArkeoNetwork/common/utils"
	"github.com/ArkeoNetwork/directory/pkg/logging"
)

type SnapshotIndexerAppParams struct {
	utils.DBConfig
}

type SnapshotIndexerApp struct {
	params SnapshotIndexerAppParams
	db     *db.AirdropDB
}

var log = logging.WithoutFields()

func NewSnapshotIndexer(params SnapshotIndexerAppParams) *SnapshotIndexerApp {
	d, err := db.New(params.DBConfig)
	if err != nil {
		panic(fmt.Sprintf("error connecting to the db: %+v", err))
	}
	return &SnapshotIndexerApp{params: params, db: d}
}

func (app *SnapshotIndexerApp) Start() {
	voters, err := app.GetSnapshotProposalVoters()
	if err != nil {
		panic(fmt.Sprintf("error getting snapshot proposal voters: %+v", err))
	}
	err = app.db.InsertVoters(voters)
	if err != nil {
		panic(fmt.Sprintf("error inserting snapshot proposal voters: %+v", err))
	}
}
