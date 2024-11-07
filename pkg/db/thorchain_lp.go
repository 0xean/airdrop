package db

import (
	"context"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (d *AirdropDB) FindLatestIndexedThorchainLPBlock(chain string, poolName string) (int64, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return 0, errors.Wrapf(err, "error obtaining db connection")
	}
	var r int64
	if err = selectOne(conn, sqlFindLatestThorchainLPBlockIndexed, &r, chain, poolName); err != nil {
		return 0, errors.Wrapf(err, "error selecting latest block")
	}
	return r, nil
}

// this writes a per block "event"
func (d *AirdropDB) InsertThorLPBalanceEvent(event []types.ThorLPBalanceEvent) error {
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
			sqlInsertThorLPBalanceEvent,
			evt.Chain,
			evt.BlockNumber,
			evt.Pool,
			evt.BalanceAsset,
			evt.BalanceRune,
			evt.AddressThor,
			evt.AddressNative,
		)
	}
	results := conn.SendBatch(context.Background(), batch)
	err = results.Close()
	if err != nil {
		return errors.Wrap(err, "error executing batch")
	}
	return nil
}
