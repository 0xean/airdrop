package db

import (
	"context"
	"strings"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (d *AirdropDB) FindStakingContracts() ([]*types.StakingContract, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining db connection")
	}
	results := make([]*types.StakingContract, 0, 128)
	if err = pgxscan.Select(context.Background(), conn, &results, sqlFindAllStakingContracts); err != nil {
		return nil, errors.Wrapf(err, "error scanning")
	}
	return results, nil
}

// find staking contracts by name
func (d *AirdropDB) FindStakingContractsByName(name string) ([]*types.StakingContract, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining db connection")
	}
	results := make([]*types.StakingContract, 0, 128)
	if err = pgxscan.Select(context.Background(), conn, &results, sqlFindStakingContractsByName, name); err != nil {
		return nil, errors.Wrapf(err, "error scanning")
	}
	return results, nil
}

func (d *AirdropDB) UpdateStakingContractHeight(stakingContractAddress string, chain string, height uint64) error {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return errors.Wrapf(err, "error obtaining db connection")
	}
	_, err = conn.Exec(context.Background(), sqlUpdateStakingContractHeight, height, strings.ToLower(stakingContractAddress), strings.ToUpper(chain))
	if err != nil {
		return errors.Wrapf(err, "error updating staking contract height")
	}

	return nil
}

func (d *AirdropDB) UpsertStakingEventBatch(stakingEvents []*types.StakingEvent) error {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return errors.Wrapf(err, "error obtaining db connection")
	}
	x := make([]interface{}, len(stakingEvents))
	for i, _ := range stakingEvents {
		x[i] = stakingEvents[i]
	}

	batch := &pgx.Batch{}
	for _, stakingEvent := range stakingEvents {
		log.Print("EVENT: ", stakingEvent)
		batch.Queue(
			sqlUpsertStakingEvent,
			stakingEvent.TxHash,
			stakingEvent.LogIndex,
			strings.ToLower(stakingEvent.Token),
			strings.ToLower(stakingEvent.StakingContract),
			strings.ToLower(stakingEvent.Staker),
			stakingEvent.Value,
			stakingEvent.BlockNumber,
			stakingEvent.Chain)
	}
	results := conn.SendBatch(context.Background(), batch)
	err = results.Close()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}
	return nil

}
