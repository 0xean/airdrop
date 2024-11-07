package cli

import (
	"github.com/ArkeoNetwork/airdrop/indexer"
	"github.com/ArkeoNetwork/common/logging"
	"github.com/ArkeoNetwork/common/utils"
	"github.com/spf13/cobra"
)

type Config struct {
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         uint   `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPass         string `mapstructure:"DB_PASS"`
	DBName         string `mapstructure:"DB_NAME"`
	DBSSLMode      string `mapstructure:"DB_SSL_MODE"`
	DBPoolMaxConns int    `mapstructure:"DB_POOL_MAX_CONNS"`
	DBPoolMinConns int    `mapstructure:"DB_POOL_MIN_CONNS"`
}

var (
	log         = logging.WithoutFields()
	indexEthCmd = &cobra.Command{
		Use:   "index-eth",
		Short: "gather eth chain data store in our db",
		Run:   runEthIndexer,
	}
)

func init() {
	indexEthCmd.Flags().BoolP("transfers", "t", false, "index token transfers")
	indexEthCmd.Flags().BoolP("rewards", "r", false, "index staking rewards")
	indexEthCmd.Flags().Bool("hedgeys", false, "index hedgey events")
}

// index ethereum things
func runEthIndexer(cmd *cobra.Command, args []string) {
	log.Info("starting data generation process")
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErr("db config undefined")
		return
	}

	indexerApp := indexer.NewIndexer(indexer.IndexerAppParams{
		DBConfig: *c,
	})
	xfers, _ := cmd.Flags().GetBool("transfers")
	rewards, _ := cmd.Flags().GetBool("rewards")
	hedgeys, _ := cmd.Flags().GetBool("hedgeys")

	if !xfers && !rewards && !hedgeys {
		log.Infof("no flags set, starting all eth indexers")
		indexerApp.Start()
		return
	}

	var err error
	if xfers {
		if err = indexerApp.IndexTransfers(); err != nil {
			cmd.PrintErrf("error indexing transfers: %+v", err)
		}
	}

	if rewards {
		if err = indexerApp.IndexStakingRewardsEvents("stakingrewardsv5"); err != nil {
			cmd.PrintErrf("error indexing staking rewards v5: %+v", err)
		}
		if err = indexerApp.IndexStakingRewardsEvents("stakingrewardsv6"); err != nil {
			cmd.PrintErrf("error indexing staking rewards v6: %+v", err)
		}
		if err = indexerApp.IndexStakingRewardsEvents("stakingrewardsv7"); err != nil {
			cmd.PrintErrf("error indexing staking rewards v7: %+v", err)
		}
		if err = indexerApp.IndexStakingRewardsEvents("stakingrewardsv8"); err != nil {
			cmd.PrintErrf("error indexing staking rewards v8: %+v", err)
		}
		if err = indexerApp.IndexStakingRewardsEvents("stakingrewardsv9"); err != nil {
			cmd.PrintErrf("error indexing staking rewards v9: %+v", err)
		}
	}

	if hedgeys {
		if err = indexerApp.IndexHedgeyEvents(); err != nil {
			cmd.PrintErrf("error indexing hedgeys: %+v", err)
		}
	}
	log.Debug("finished data generation process")
}
