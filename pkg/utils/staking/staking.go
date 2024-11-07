package staking

import (
	"context"
	"errors"

	"github.com/ArkeoNetwork/common/logging"

	"github.com/ArkeoNetwork/airdrop/contracts/stakingrewards"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

var log = logging.WithoutFields()

func GetAllStakedEvents(
	startBlock uint64,
	endBlock uint64,
	batchSize uint64,
	staking *stakingrewards.Stakingrewards) (*[]*stakingrewards.StakingrewardsStaked, error) {
	events := []*stakingrewards.StakingrewardsStaked{}
	currentBlock := startBlock
	retryCount := 20
	for currentBlock < endBlock {
		end := currentBlock + batchSize
		filterOpts := bind.FilterOpts{
			Start:   currentBlock,
			End:     &end,
			Context: context.Background(),
		}
		iter, err := staking.FilterStaked(&filterOpts, nil)
		if err != nil {
			log.Errorf("failed to get transfer events for block %+v retring", err)
			retryCount--
			if retryCount < 0 {
				return nil, errors.New("GetAllTransfers failed with 0 retries")
			}
			continue
		}
		for iter.Next() {
			events = append(events, iter.Event)
		}
		currentBlock = end
	}
	return &events, nil
}

func GetAllWithdrawnEvents(
	startBlock uint64,
	endBlock uint64,
	batchSize uint64,
	staking *stakingrewards.Stakingrewards) (*[]*stakingrewards.StakingrewardsWithdrawn, error) {
	events := []*stakingrewards.StakingrewardsWithdrawn{}
	currentBlock := startBlock
	retryCount := 20
	for currentBlock < endBlock {
		end := currentBlock + batchSize
		filterOpts := bind.FilterOpts{
			Start:   currentBlock,
			End:     &end,
			Context: context.Background(),
		}
		iter, err := staking.FilterWithdrawn(&filterOpts, nil)
		if err != nil {
			log.Errorf("failed to get transfer events for block %+v retring", err)
			retryCount--
			if retryCount < 0 {
				return nil, errors.New("GetAllTransfers failed with 0 retries")
			}
			continue
		}
		for iter.Next() {
			events = append(events, iter.Event)
		}
		currentBlock = end
	}
	return &events, nil
}
