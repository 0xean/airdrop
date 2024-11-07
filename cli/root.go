package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "arkeodrop",
		Short: "arkeodrop is airdrop utilities",
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			println("arkeodrop v0.0.1\n")
		},
	}

	indexCosmosCmd = &cobra.Command{
		Use:   "index-cosmos",
		Short: "cosmos-sdk indexing",
	}

	snapshotIndexCmd = &cobra.Command{
		Use:   "snapshot-index",
		Short: "gather snapshot data store in our db",
		Run:   runSnapshotIndexer,
	}
)

func init() {
	rootCmd.PersistentFlags().StringP("env", "e", "docker/dev/docker.env", "env file to source")
	rootCmd.AddCommand(versionCmd)
	// <<<<<<< HEAD
	// rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(snapshotIndexCmd)
	// =======
	rootCmd.AddCommand(indexEthCmd)
	indexCosmosCmd.AddCommand(indexDelegatorsCmd)
	indexCosmosCmd.AddCommand(indexThorchainLPCmd)
	indexCosmosCmd.AddCommand(delegationsFromStateExport)
	indexCosmosCmd.AddCommand(liquidityFromStateExport)
	rootCmd.AddCommand(indexCosmosCmd)
	// >>>>>>> 845abdb (cosmos staking/lp)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(exportFarmCmd)
	rootCmd.AddCommand(exportDelegatesCmd)
	rootCmd.AddCommand(exportOsmoLpCmd)
	rootCmd.AddCommand(exportThorchainLPCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
