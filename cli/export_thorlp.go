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
	exportThorchainLPCmd = &cobra.Command{
		Use:   "export-thorlp [pool]",
		Short: "export aggregate thorchain lp data for a given pool",
		Run:   runExportThorchainLPAvg,
		Args:  cobra.ExactValidArgs(1),
	}
)

// export thorchain lp data for a given pool
func runExportThorchainLPAvg(cmd *cobra.Command, args []string) {
	log.Infof("starting ThorchainLP export process for %s", args[0])
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
		fileName = fmt.Sprintf("/tmp/airdrop_thorlp_%s.csv", args[0])
	}

	err := exportThorchainLPAvgs(*c, args[0], fileName)
	if err != nil {
		log.Errorf("error exporting: %+v", err)
	}
}

func exportThorchainLPAvgs(dbConfig utils.DBConfig, pool, fileName string) error {
	d, err := db.New(dbConfig)
	if err != nil {
		return errors.Wrapf(err, "error connecting to the db")
	}

	avgs, err := d.FindAveragedThorLPBalances(pool)
	if err != nil {
		return errors.Wrapf(err, "error finding averages")
	}
	log.Debugf("found %d thorlp averages for %s", len(avgs), pool)
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
