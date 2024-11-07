package db

import (
	"testing"

	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	arkutils "github.com/ArkeoNetwork/common/utils"
	"github.com/stretchr/testify/assert"
)

// #javanaming
func TestFindLatestIndexedCosmosStakingBlock(t *testing.T) {
	c := arkutils.ReadDBConfig(utils.GetEnvPath())
	if !assert.NotNil(t, c) {
		t.FailNow()
	}
	d, err := New(*c)
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	if !assert.NotNil(t, d) {
		t.FailNow()
	}

	latest, err := d.FindLatestIndexedCosmosStakingBlock("GAIA")
	if !assert.Nil(t, err) {
		t.FailNow()
	}
	if assert.NotNil(t, latest) {
		assert.Greater(t, latest, int64(0))
	}
}
