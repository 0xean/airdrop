package indexer

import (
	"context"
	"time"

	"github.com/ArkeoNetwork/airdrop/contracts/erc20"
	"github.com/ArkeoNetwork/airdrop/contracts/stakingrewards"
	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

func (app *IndexerApp) IndexStakingRewardsEvents(contractName string) error {
	// get all staking contracts
	stakingContracts, err := app.db.FindStakingContractsByName(contractName)
	if err != nil {
		log.Info("error finding all staking contracts")
		return errors.Wrap(err, "error finding all staking contracts")
	}
	// for each staking contract
	for _, stakingContract := range stakingContracts {
		log.Info("indexing staking contract: ", stakingContract.ContractName)
		// get chain info and create rpc client
		chain, err := app.db.FindChain(stakingContract.Chain)
		if err != nil {
			return errors.Wrap(err, "error finding chain for staking contract")
		}
		client, err := ethclient.Dial(chain.RpcUrl)
		if err != nil {
			return errors.Wrapf(err, "failed to connect to eth RPC client %s", chain.RpcUrl)
		}
		// create staking contract
		staking, err := stakingrewards.NewStakingrewards(common.HexToAddress(stakingContract.Address), client)
		if err != nil {
			return errors.Wrap(err, "error creating staking contract")
		}

		// get the token address and determine decimals
		stakingTokenAddress, err := staking.StakingToken(nil)
		if err != nil {
			return errors.Wrap(err, "error getting staking token address")
		}
		stakingToken, err := erc20.NewErc20(stakingTokenAddress, client)
		if err != nil {
			return errors.Wrap(err, "error creating staking token contract")
		}
		stakingTokenDecimals, err := stakingToken.Decimals(nil)
		if err != nil {
			return errors.Wrap(err, "error getting staking token decimals")
		}

		startBlock := stakingContract.Height
		if startBlock < stakingContract.GenesisBlock {
			startBlock = stakingContract.GenesisBlock
		}

		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			return errors.Wrap(err, "failed to get current block number from RPC client")
		}
		endBlock := chain.SnapshotEndBlock
		if blockNumber < endBlock {
			endBlock = blockNumber
		}
		log.Infof("Connected to client for %s. Current block %d Indexing staking events from block %d", chain.Name, blockNumber, startBlock)
		err = app.indexStakingRewardContractEvents(
			startBlock,
			endBlock,
			1000,
			chain.Name,
			stakingTokenDecimals,
			stakingTokenAddress.String(),
			stakingContract.Address,
			staking)
		if err != nil {
			return errors.Wrap(err, "error indexing staking contract events")
		}
	}
	return nil
}

func (app *IndexerApp) indexStakingRewardContractEvents(
	startBlock uint64,
	endBlock uint64,
	batchSize uint64,
	chain string,
	stakingTokenDecimals uint8,
	stakingTokenAddress string,
	stakingContractAddress string,
	stakingContract *stakingrewards.Stakingrewards) error {
	currentBlock := startBlock
	retryCount := 20
	for currentBlock < endBlock {
		end := currentBlock + batchSize
		filterOpts := bind.FilterOpts{
			Start:   currentBlock,
			End:     &end,
			Context: context.Background(),
		}
		// handle staked events
		iter, err := stakingContract.FilterStaked(&filterOpts, nil)
		if err != nil {
			log.Errorf("failed to get staked events for block %+v retrying", err)
			retryCount--
			if retryCount < 0 {
				return errors.New("indexStakingContractEvents failed with 0 retries")
			}
			continue
		}

		stakingEvents := []*types.StakingEvent{}
		for iter.Next() {
			stakingValueDecimal := utils.BigIntToFloat(iter.Event.Amount, stakingTokenDecimals)
			stakingEvents = append(stakingEvents,
				&types.StakingEvent{
					LogIndex:        iter.Event.Raw.Index,
					Value:           stakingValueDecimal,
					BlockNumber:     iter.Event.Raw.BlockNumber,
					TxHash:          iter.Event.Raw.TxHash.String(),
					StakingContract: stakingContractAddress,
					Staker:          iter.Event.User.String(),
					Token:           stakingTokenAddress,
					Chain:           chain,
				})
		}

		// handle unstaked events
		iterWithdrawn, err := stakingContract.FilterWithdrawn(&filterOpts, nil)
		if err != nil {
			log.Errorf("failed to get staked events for block %+v retrying", err)
			retryCount--
			if retryCount < 0 {
				return errors.New("indexStakingContractEvents failed with 0 retries")
			}
			continue
		}

		for iterWithdrawn.Next() {
			stakingValueDecimal := utils.BigIntToFloat(iterWithdrawn.Event.Amount, stakingTokenDecimals) * -1 //negative value for unstaked
			stakingEvents = append(stakingEvents,
				&types.StakingEvent{
					LogIndex:        iterWithdrawn.Event.Raw.Index,
					Value:           stakingValueDecimal,
					BlockNumber:     iterWithdrawn.Event.Raw.BlockNumber,
					TxHash:          iterWithdrawn.Event.Raw.TxHash.String(),
					StakingContract: stakingContractAddress,
					Staker:          iterWithdrawn.Event.User.String(),
					Token:           stakingTokenAddress,
					Chain:           chain,
				})
		}

		err = app.db.UpdateStakingContractHeight(stakingContractAddress, chain, end)
		if err != nil {
			log.Warnf("failed to update Staking Contract height %+v", err)
		}
		currentBlock = end

		if len(stakingEvents) == 0 {
			continue
		}

		err = app.db.UpsertStakingEventBatch(stakingEvents)
		if err != nil {
			log.Errorf("failed to upsert staking event batch %+v", err)
			return err
		}
		log.Debugf("Updated staking events for blocks through %d with %d events", end, len(stakingEvents))
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
