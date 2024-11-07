package db

import (
	"context"
	"strings"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/pkg/errors"
)

// findallchains - queries chains table for all chains
func (d *AirdropDB) FindAllChains() ([]*types.Chain, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining db connection")
	}
	results := make([]*types.Chain, 0, 10)
	if err = pgxscan.Select(context.Background(), conn, &results, sqlFindAllChains); err != nil {
		return nil, errors.Wrapf(err, "error scanning")
	}
	return results, nil
}

func (d *AirdropDB) FindETHChains() ([]*types.Chain, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining db connection")
	}
	results := make([]*types.Chain, 0, 10)
	if err = pgxscan.Select(context.Background(), conn, &results, sqlFindEthChains); err != nil {
		return nil, errors.Wrapf(err, "error scanning")
	}
	return results, nil
}

// findchain - queries chains table for a chain
func (d *AirdropDB) FindChain(chain string) (*types.Chain, error) {
	conn, err := d.getConnection()
	defer conn.Release()
	if err != nil {
		return nil, errors.Wrapf(err, "error obtaining db connection")
	}
	result := &types.Chain{}
	if err = pgxscan.Get(context.Background(), conn, result, sqlFindChain, strings.ToUpper(chain)); err != nil {
		return nil, errors.Wrapf(err, "error scanning")
	}
	return result, nil
}
