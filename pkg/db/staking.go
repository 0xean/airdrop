package db

import (
	"context"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// findStakingContracts
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

// update staking contract height
func (d *AirdropDB) UpdateStakingContractHeight(stakingContractAddress string, height uint64) error {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return errors.Wrapf(err, "error obtaining db connection")
	}
	_, err = conn.Exec(context.Background(), sqlUpdateStakingContractHeight, height, stakingContractAddress)
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
		batch.Queue(
			sqlUpsertStakingEvent,
			stakingEvent.TxHash,
			stakingEvent.LogIndex,
			stakingEvent.Token,
			stakingEvent.StakingContract,
			stakingEvent.Staker,
			stakingEvent.Value,
			stakingEvent.BlockNumber)
	}
	results := conn.SendBatch(context.Background(), batch)
	err = results.Close()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}
	return nil

}
