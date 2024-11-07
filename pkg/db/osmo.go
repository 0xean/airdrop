package db

import (
	"context"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (d *AirdropDB) InsertOsmoLP(event []*types.OsmoLP) error {
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
			sqlInsertOsmoLP,
			evt.BlockNumber,
			evt.Account,
			evt.LpAmount,
			evt.PoolId,
			evt.TxHash,
			evt.Type,
		)
	}
	results := conn.SendBatch(context.Background(), batch)
	err = results.Close()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}
	return nil
}
