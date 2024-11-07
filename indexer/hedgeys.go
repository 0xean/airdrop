package indexer

import (
	"context"
	"strings"

	"github.com/ArkeoNetwork/airdrop/contracts/erc20"
	"github.com/ArkeoNetwork/airdrop/contracts/hedgey"
	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

func (app *IndexerApp) IndexHedgeyEvents() error {
	// get all staking contracts
	stakingContracts, err := app.db.FindStakingContractsByName("hedgeyNFT")
	if err != nil {
		return errors.Wrap(err, "error finding all hedgeyNFT contracts")
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
		// create hedgey contract
		hedgeyContract, err := hedgey.NewHedgey(common.HexToAddress(stakingContract.Address), client)
		if err != nil {
			return errors.Wrap(err, "error creating hedgey contract")
		}

		// get the token address and determine decimals
		foxToken, err := app.db.FindTokenByChainAndSymbol(stakingContract.Chain, "FOX")
		if err != nil {
			return errors.Wrap(err, "error getting FOX token address")
		}
		stakingToken, err := erc20.NewErc20(common.HexToAddress(foxToken.Address), client)
		if err != nil {
			return errors.Wrap(err, "error creating staking token contract")
		}
		foxTokenDecimals, err := stakingToken.Decimals(nil)
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
		log.Infof("Connected to client for %s. Current block %d Indexing hedgey events from block %d", chain.Name, blockNumber, startBlock)
		err = app.indexHedgeyContractEvents(
			startBlock,
			endBlock,
			1000,
			chain.Name,
			foxTokenDecimals,
			foxToken.Address,
			stakingContract.Address,
			hedgeyContract)
		if err != nil {
			return errors.Wrap(err, "error indexing hedgey contract events")
		}
	}
	log.Infof("Finished Indexing Hedgey Events")
	return nil
}

func (app *IndexerApp) indexHedgeyContractEvents(
	startBlock uint64,
	endBlock uint64,
	batchSize uint64,
	chain string,
	foxTokenDecimals uint8,
	foxTokenAddress string,
	hedgeyContractAddress string,
	hedgeyContract *hedgey.Hedgey) error {
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
		iter, err := hedgeyContract.FilterNFTCreated(&filterOpts)
		if err != nil {
			log.Errorf("failed to get staked events for block %+v retrying", err)
			retryCount--
			if retryCount < 0 {
				return errors.New("indexHedgeyContractEvents failed with 0 retries")
			}
			continue
		}

		stakingEvents := []*types.StakingEvent{}
		for iter.Next() {
			// confirm these are for the correct token
			if strings.ToLower(iter.Event.Token.String()) != foxTokenAddress {
				continue
			}

			stakingValueDecimal := utils.BigIntToFloat(iter.Event.Amount, foxTokenDecimals)
			stakingEvents = append(stakingEvents,
				&types.StakingEvent{
					LogIndex:        iter.Event.Raw.Index,
					Value:           stakingValueDecimal,
					BlockNumber:     iter.Event.Raw.BlockNumber,
					TxHash:          iter.Event.Raw.TxHash.String(),
					StakingContract: hedgeyContractAddress,
					Staker:          iter.Event.Holder.String(),
					Token:           foxTokenAddress,
					Chain:           chain,
				})
		}

		// handle unstaked events
		iterWithdrawn, err := hedgeyContract.FilterNFTRedeemed(&filterOpts)
		if err != nil {
			log.Errorf("failed to get staked events for block %+v retrying", err)
			retryCount--
			if retryCount < 0 {
				return errors.New("indexStakingContractEvents failed with 0 retries")
			}
			continue
		}

		for iterWithdrawn.Next() {
			// confirm these are for the correct token
			if strings.ToLower(iterWithdrawn.Event.Token.String()) != foxTokenAddress {
				continue
			}
			stakingValueDecimal := utils.BigIntToFloat(iterWithdrawn.Event.Amount, foxTokenDecimals) * -1 //negative value for unstaked
			stakingEvents = append(stakingEvents,
				&types.StakingEvent{
					LogIndex:        iterWithdrawn.Event.Raw.Index,
					Value:           stakingValueDecimal,
					BlockNumber:     iterWithdrawn.Event.Raw.BlockNumber,
					TxHash:          iterWithdrawn.Event.Raw.TxHash.String(),
					StakingContract: hedgeyContractAddress,
					Staker:          iterWithdrawn.Event.Holder.String(),
					Token:           foxTokenAddress,
					Chain:           chain,
				})
		}

		err = app.db.UpdateStakingContractHeight(hedgeyContractAddress, chain, end)
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
		log.Debugf("Updated hedgey events for blocks through %d with %d events", end, len(stakingEvents))
	}
	return nil
}
