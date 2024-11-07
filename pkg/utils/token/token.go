package token

import (
	"math/big"
	"sync"

	"github.com/ArkeoNetwork/common/logging"

	"github.com/ArkeoNetwork/airdrop/contracts/erc20"

	"github.com/ethereum/go-ethereum/common"
)

var log = logging.WithoutFields()

func GetAllHolders(transferEvents *[]*erc20.Erc20Transfer) *[]common.Address {
	isHolder := map[common.Address]bool{}
	for _, transfer := range *transferEvents {
		isHolder[transfer.To] = true
	}
	holders := make([]common.Address, len(isHolder))

	i := 0
	for k := range isHolder {
		holders[i] = k
		i++
	}

	return &holders
}

func mapTransfersByAddress(transferEvents *[]*erc20.Erc20Transfer) *map[common.Address]*[]*erc20.Erc20Transfer {
	transfersByAddress := map[common.Address]*[]*erc20.Erc20Transfer{}
	for _, transfer := range *transferEvents {
		addTransferToAddressMap(&transfersByAddress, transfer, transfer.From)
		addTransferToAddressMap(&transfersByAddress, transfer, transfer.To)
	}
	return &transfersByAddress
}

func addTransferToAddressMap(transfersByAddress *map[common.Address]*[]*erc20.Erc20Transfer, transfer *erc20.Erc20Transfer, address common.Address) {
	transfersByAddressLocal := *transfersByAddress
	transfers, exists := transfersByAddressLocal[address]
	if !exists {
		transfers := []*erc20.Erc20Transfer{transfer}
		transfersByAddressLocal[address] = &transfers
	} else {
		updatedTransfers := append(*transfers, transfer)
		transfersByAddressLocal[address] = &updatedTransfers
	}
}

func GetBalancesAtBlock(allHolders *[]common.Address, transferEvents *[]*erc20.Erc20Transfer, block uint64) *map[common.Address]*big.Int {
	transfersByAddress := mapTransfersByAddress(transferEvents)
	balancesAtBlockByAddress := map[common.Address]*big.Int{}
	for _, address := range *allHolders {
		currentBalance := big.NewInt(0)
		// get all events for this address
		transferEvents := (*transfersByAddress)[address]
		for _, transferEvent := range *transferEvents {
			if transferEvent.Raw.BlockNumber > block {
				break
			}

			// update the users current balance
			if transferEvent.From == address {
				currentBalance.Sub(currentBalance, transferEvent.Value)
			} else if transferEvent.To == address {
				currentBalance.Add(currentBalance, transferEvent.Value)
			} else {
				panic("bad address")
			}

		}
		balancesAtBlockByAddress[address] = currentBalance
	}
	return &balancesAtBlockByAddress
}

func GenerateBalanceHistory(
	allHolders *[]common.Address,
	transferEvents *[]*erc20.Erc20Transfer,
	startBlock uint64,
	endBlock uint64,
) *map[common.Address]*[]*big.Int {
	transfersByAddress := mapTransfersByAddress(transferEvents)

	balanceHistory := map[common.Address]*[]*big.Int{}
	blockCount := endBlock - startBlock

	for i, address := range *allHolders {
		var balances []*big.Int = make([]*big.Int, blockCount)
		balanceHistory[address] = &balances
		currentBlock := startBlock
		var currentBalance *big.Int = big.NewInt(0)

		// get all events for this address
		transferEvents := (*transfersByAddress)[address]
		log.Debugf("Processing %d of %d : %s with %d transfer events", i, len(*allHolders), address.String(), len(*transferEvents))
		for eventCounter, transferEvent := range *transferEvents {
			if transferEvent.Raw.BlockNumber > endBlock {
				// event is after the window, we are done with this address
				break
			}

			if transferEvent.Raw.BlockNumber < startBlock {
				// event is before window, we need to update balance to track

				// update the users current balance
				currentBalance = addEventToBalance(transferEvent, address, currentBalance)

				// we also need to be careful to fill forward if no other events
				if eventCounter == len(*transferEvents)-1 {
					currentBlock = fillBalanceForward(startBlock, endBlock, currentBlock, currentBalance, &balances)
				}
				continue
			}

			// event is during of after window, we need to fill previous balance before we udpate it
			for transferEvent.Raw.BlockNumber > currentBlock && currentBlock < endBlock {
				// fill in users current balance from the last block we saw to
				// this event
				balances[currentBlock-startBlock] = currentBalance
				currentBlock++
			}

			// we can now update the users current balance
			currentBalance = addEventToBalance(transferEvent, address, currentBalance)
			balances[currentBlock-startBlock] = currentBalance
			currentBlock++

			if eventCounter == len(*transferEvents)-1 {
				// this is the last event for this user, fill forward their balance
				currentBlock = fillBalanceForward(startBlock, endBlock, currentBlock, currentBalance, &balances)
			}

		}
	}
	return &balanceHistory
}

func addEventToBalance(transferEvent *erc20.Erc20Transfer, address common.Address, currentBalance *big.Int) *big.Int {
	// note: this creates a deep copy of the big into to return, because we don't want to corrupt the previous
	// values that are referenced by pointer in our balance arrays
	if transferEvent.From == address {
		currentBalance = new(big.Int).Set(currentBalance)
		return currentBalance.Sub(currentBalance, transferEvent.Value)
	} else if transferEvent.To == address {
		currentBalance = new(big.Int).Set(currentBalance)
		return currentBalance.Add(currentBalance, transferEvent.Value)
	} else {
		panic("bad address")
	}
}

func GetBlockWeigthedAverageBalance(balanceHistory *map[common.Address]*[]*big.Int) *map[common.Address]*big.Int {
	averageBalanceByAddress := map[common.Address]*big.Int{}
	mutex := sync.Mutex{}
	var waitGroup sync.WaitGroup
	for addressKey, balancesVal := range *balanceHistory {
		waitGroup.Add(1)
		go func(address common.Address, balances []*big.Int, avgBalanceByAddress *map[common.Address]*big.Int) {
			defer waitGroup.Done()
			balance := getAverageBalance(&balances)
			mutex.Lock() // go cannot handle concurrent writes to different keys in the map.
			(*avgBalanceByAddress)[address] = balance
			mutex.Unlock()
		}(addressKey, *balancesVal, &averageBalanceByAddress)
	}
	waitGroup.Wait()
	return &averageBalanceByAddress
}

func getAverageBalance(balances *[]*big.Int) *big.Int {
	balanceSum := big.NewInt(0)
	for _, balance := range *balances {
		balanceSum.Add(balanceSum, balance)
	}
	return balanceSum.Div(balanceSum, big.NewInt(int64(len(*balances))))
}

func fillBalanceForward(startBlock uint64, endBlock uint64, currentBlock uint64, currentBalance *big.Int, balances *[]*big.Int) uint64 {
	// this is the last event for this user, fill forward their balance
	for currentBlock < endBlock {
		(*balances)[currentBlock-startBlock] = currentBalance
		currentBlock++
	}
	return currentBlock
}
