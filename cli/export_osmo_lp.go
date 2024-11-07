package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ArkeoNetwork/airdrop/pkg/db"
	"github.com/ArkeoNetwork/common/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	exportOsmoLpCmd = &cobra.Command{
		Use:   "export-lp [chain]",
		Short: "export aggregate delegate data",
		Run:   runExportOsmoLpAvg,
		Args:  cobra.ExactValidArgs(1),
	}
)

func init() {
	exportOsmoLpCmd.Flags().StringP("output", "f", "", "csv output file, default /tmp/airdrop_{chain}.csv")
}

// export block weighted delegation averages for given cosmos chain and validator
func runExportOsmoLpAvg(cmd *cobra.Command, args []string) {
	log.Infof("starting delegate export process for %s", args[0])
	flags := cmd.InheritedFlags()
	envPath, _ := flags.GetString("env")
	c := utils.ReadDBConfig(envPath)
	if c == nil {
		cmd.PrintErrln("db config undefined")
		return
	}
	flags = cmd.Flags()
	fileName, _ := flags.GetString("output")
	if fileName == "" {
		fileName = fmt.Sprintf("/tmp/airdrop_%s_osmo_lp.csv", args[0])
	}

	err := exportWeightedOsmoLpAvgs(*c, args[0], fileName)
	if err != nil {
		log.Errorf("error exporting: %+v", err)
	}
}

func exportWeightedOsmoLpAvgs(dbConfig utils.DBConfig, chain, fileName string) error {
	d, err := db.New(dbConfig)
	if err != nil {
		return errors.Wrapf(err, "error connecting to the db")
	}

	avgs, err := d.FindAveragedOsmoLpBalances(chain)
	if err != nil {
		return errors.Wrapf(err, "error finding averages")
	}
	log.Debugf("found %d delegation averages for %s", len(avgs), chain)
	sb := strings.Builder{}
	fmt.Fprint(&sb, "address,balance\n")
	for _, a := range avgs {
		fmt.Fprintf(&sb, "\"%s\",%.18f\n", a.Address, a.Holding)
	}

	if err = os.WriteFile(fileName, []byte(sb.String()), os.ModePerm); err != nil {
		return errors.Wrapf(err, "error writing %s", fileName)
	}
	fmt.Printf("wrote %s\n", fileName)
	return nil
}
