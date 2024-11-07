// test for chains.go
package db

import (
	"fmt"
	"testing"

	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	arkutils "github.com/ArkeoNetwork/common/utils"
	"github.com/stretchr/testify/assert"
)

// TestFindAllChains - tests FindAllChains
func TestFindAllChains(t *testing.T) {
	db, err := New(config)
	if err != nil {
		t.Errorf("error getting db: %+v", err)
	}

	// test
	chains, err := db.FindAllChains()
	if err != nil {
		t.Fatalf("unable to find chains for tokens: %+v", err)
	}
	// verify
	if len(chains) < 1 {
		t.Fatalf("expected > 1 chains, found %d", len(chains))
	}
}

func TestFindChain(t *testing.T) {
	c := arkutils.ReadDBConfig(utils.GetEnvPath())
	if c == nil {
		fmt.Print("error: no config loaded")
		return
	}
	db, err := New(*c)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	chain, err := db.FindChain("GAIA")
	assert.Nil(t, err)
	if assert.NotNil(t, chain) {
		assert.Equal(t, uint8(6), chain.Decimals)
	}

}
