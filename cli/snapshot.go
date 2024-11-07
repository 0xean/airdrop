package cli

import (
	"github.com/ArkeoNetwork/airdrop/snapshot"
	"github.com/ArkeoNetwork/common/utils"
	"github.com/spf13/cobra"
)

func runSnapshotIndexer(cmd *cobra.Command, args []string) {
	log.Info("starting gethering snapshot data process")
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErrf("no config for path %s", envPath)
		return
	}
	snapshotIndexerApp := snapshot.NewSnapshotIndexer(snapshot.SnapshotIndexerAppParams{
		DBConfig: utils.DBConfig{
			DBHost:         c.DBHost,
			DBPort:         c.DBPort,
			DBUser:         c.DBUser,
			DBPass:         c.DBPass,
			DBName:         c.DBName,
			DBPoolMaxConns: c.DBPoolMaxConns,
			DBPoolMinConns: c.DBPoolMinConns,
			DBSSLMode:      c.DBSSLMode,
		},
	})
	snapshotIndexerApp.Start()
	log.Debug("finished gethering snapshot data process")
}
