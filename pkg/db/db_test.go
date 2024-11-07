package db

import (
	"context"
	"testing"

	"github.com/ArkeoNetwork/common/utils"
)

var config = utils.DBConfig{
	DBHost:         "localhost",
	DBPort:         5432,
	DBUser:         "arkeo",
	DBPass:         "arkeo123",
	DBName:         "arkeo_airdrop",
	DBPoolMaxConns: 2,
	DBPoolMinConns: 1,
	DBSSLMode:      "prefer",
}

func TestNew(t *testing.T) {

	db, err := New(config)
	if err != nil {
		t.Errorf("error: %+v", err)
	}
	conn, err := db.pool.Acquire(context.Background())
	if err != nil {
		t.Errorf("error acquiring connection: %+v", err)
	}
	defer conn.Release()
	log.Infof("got connection %s", conn)
}
