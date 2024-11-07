package db

import (
	"context"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (d *AirdropDB) InsertStakingEvents(event []*types.CosmosStakingEvent) error {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return errors.Wrapf(err, "error obtaining db connection")
	}
	x := make([]interface{}, len(event))
	for i := range event {
		x[i] = event[i]
	}

	batch := &pgx.Batch{}
	for _, evt := range event {
		batch.Queue(
			sqlInsertCosmosStakingEvent,
			evt.Chain,
			evt.EventType,
			evt.Delegator,
			evt.Validator,
			evt.Value,
			evt.BlockNumber,
			evt.TxHash,
			evt.EventIndex,
		)
	}
	results := conn.SendBatch(context.Background(), batch)
	err = results.Close()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}
	return nil
}

func (d *AirdropDB) FindLatestIndexedCosmosStakingBlock(chain string) (int64, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return 0, errors.Wrapf(err, "error obtaining db connection")
	}
	var r int64
	if err = selectOne(conn, sqlFindLatestCosmosStakingBlockIndexed, &r, chain); err != nil {
		return 0, errors.Wrapf(err, "error selecting latest block")
	}
	return r, nil
}

func (d *AirdropDB) FindLatestIndexedCosmosLPBlock(chain string) (int64, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return 0, errors.Wrapf(err, "error obtaining db connection")
	}
	var r int64
	if err = selectOne(conn, sqlFindLatestCosmosStakingBlockIndexed, &r, chain); err != nil {
		return 0, errors.Wrapf(err, "error selecting latest block")
	}
	return r, nil
}
