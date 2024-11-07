package cli

import (
	"fmt"
	"strconv"

	"github.com/ArkeoNetwork/airdrop/indexer"
	"github.com/ArkeoNetwork/common/utils"
	"github.com/spf13/cobra"
)

var (
	delegationsFromStateExport = &cobra.Command{
		Use:   "import-delegations [data-directory] [chain] [height]",
		Short: "import staked delegations from a cosmos daemond export",
		Run:   runDelegationsFromStateExport,
		Args:  cobra.ExactValidArgs(3),
	}
	liquidityFromStateExport = &cobra.Command{
		Use:   "import-liquidity [data-directory] [chain] [height]",
		Short: "import liquidity from a cosmos daemond export",
		Run:   runLiquidityFromStateExport,
		Args:  cobra.ExactValidArgs(3),
	}
	indexDelegatorsCmd = &cobra.Command{
		Use:   "delegators [chain]",
		Short: "gather cosmos-sdk chain data store in our db",
		Run:   runDelegatorsIndexer,
		Args:  cobra.ExactValidArgs(1),
	}
	indexThorchainLPCmd = &cobra.Command{
		Use:   "thor-lp [pool]",
		Short: "gather cosmos-sdk chain liquidity provider data store in our db",
		Run:   runThorLPIndexer,
		Args:  cobra.ExactValidArgs(1),
	}
)

func runDelegatorsIndexer(cmd *cobra.Command, args []string) {
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		fmt.Print("error: no config loaded")
		return
	}

	chain := args[0]
	params := indexer.CosmosIndexerParams{Chain: chain, DB: *c}
	indxr, err := indexer.NewCosmosIndexer(params)
	if err != nil {
		cmd.PrintErrf("error creating cosmos indexer: %+v", err)
		return
	}

	if err := indxr.IndexCosmosDelegators(); err != nil {
		fmt.Printf("error indexing delegators: %+v", err)
		cmd.PrintErrf("error indexing delegators: %+v", err)
		return
	}
}

func runThorSaversIndexer(cmd *cobra.Command, args []string) {
	// TODO - this is identical to thor-lp with exception of the endpoint
}

func runThorLPIndexer(cmd *cobra.Command, args []string) {
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErrf("no config for path %s", envPath)
		return
	}

	poolName := args[0]
	params := indexer.CosmosIndexerParams{Chain: "THOR", DB: *c}
	indxr, err := indexer.NewCosmosIndexer(params)
	if err != nil {
		cmd.PrintErrf("error creating cosmos indexer: %+v", err)
		return
	}

	if err = indxr.IndexThorLP(poolName); err != nil {
		cmd.PrintErrf("error indexing LP: %+v", err)
	}
}

func runDelegationsFromStateExport(cmd *cobra.Command, args []string) {
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErrf("no config for path %s", envPath)
		return
	}

	chain := args[1]
	baseDataDir := args[0]
	sheight := args[2]
	height, err := strconv.ParseInt(sheight, 10, 64)
	if err != nil {
		cmd.PrintErrf("error parsing height %s: %+v", sheight, err)
		return
	}

	params := indexer.CosmosIndexerParams{Chain: chain, DB: *c}
	indxr, err := indexer.NewCosmosIndexer(params)
	if err != nil {
		cmd.PrintErrf("error creating cosmos indexer: %+v", err)
		return
	}
	if err = indxr.IndexDelegationsFromStateExport(baseDataDir, chain, height); err != nil {
		cmd.PrintErrf("error indexing Delegations from state export: %+v", err)
	}
}


func runLiquidityFromStateExport(cmd *cobra.Command, args []string) {
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErrf("no config for path %s", envPath)
		return
	}

	chain := args[1]
	baseDataDir := args[0]
	sheight := args[2]
	height, err := strconv.ParseInt(sheight, 10, 64)
	if err != nil {
		cmd.PrintErrf("error parsing height %s: %+v", sheight, err)
		return
	}

	params := indexer.CosmosIndexerParams{Chain: chain, DB: *c}
	indxr, err := indexer.NewCosmosIndexer(params)
	if err != nil {
		cmd.PrintErrf("error creating cosmos indexer: %+v", err)
		return
	}
	if err = indxr.IndexLiquidityFromStateExport(baseDataDir, chain, height); err != nil {
		cmd.PrintErrf("error indexing Delegations from state export: %+v", err)
	}
}
