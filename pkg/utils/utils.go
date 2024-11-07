package utils

import (
	"math"
	"math/big"
	"os"

	"github.com/ArkeoNetwork/common/logging"
)

var log = logging.WithoutFields()

func BigIntToFloat(value *big.Int, decimals uint8) float64 {
	transferValue := new(big.Float).SetInt(value)
	transferValue.Quo(transferValue, big.NewFloat(float64(math.Pow10(int(decimals)))))
	transferValueDecimal, _ := transferValue.Float64()
	return transferValueDecimal
}

// Not safe for large precisions (ETH)
func FromBaseUnits(in int64, precision uint8) float64 {
	return float64(in) / math.Pow10(int(precision))
}

func GetEnvPath() string {
	envPath := os.Getenv("AIRDROP_ENV_PATH")
	if envPath == "" {
		log.Warn("no value in $AIRDROP_ENV_PATH")
	}

	return envPath
}
