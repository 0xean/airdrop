package indexer

import (
	"context"
	"time"

	"github.com/ArkeoNetwork/airdrop/contracts/erc20"
	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ArkeoNetwork/airdrop/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

func (app *IndexerApp) IndexTransfers() error {
	// find all chains we care about
	chains, err := app.db.FindETHChains()
	if err != nil {
		return errors.Wrap(err, "unable to find chains for tokens")
	}

	// todo: kick off new go-routine for each chain
	for _, chain := range chains {
		client, err := ethclient.Dial(chain.RpcUrl)
		if err != nil {
			return errors.Wrap(err, "failed to connect to eth RPC client")
		}

		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			return errors.Wrap(err, "failed to get current block number from RPC client")
		}

		log.Infof("Connected to client for %s. Current block %d", chain.Name, blockNumber)
		snapshotEnd := blockNumber
		if blockNumber > chain.SnapshotEndBlock {
			snapshotEnd = chain.SnapshotEndBlock
		}

		// get the tokens for each chain
		tokens, err := app.db.FindTokensByChain(chain.Name)
		if err != nil {
			return errors.Wrap(err, "unbale to find tokens for chain")
		}

		// iterate tokens array and get the transfers for each token
		for _, token := range tokens {
			// determine if the token has been synced to a different block
			startBlock := token.GenesisBlock
			if token.Height > startBlock {
				startBlock = token.Height
			}

			// get the transfers for each token
			log.Infof("Getting transfers for token: %s from block: %d to block: %d", token.Name, startBlock, snapshotEnd)
			tokenContract, err := erc20.NewErc20(common.HexToAddress(token.Address), client)
			if err != nil {
				return errors.Wrap(err, "unbale to get token contract for token")
			}

			err = app.indexTransfersForToken(startBlock, snapshotEnd, 1000, token.Address, tokenContract)
			if err != nil {
				return errors.Wrap(err, "unable to get transfers for token")
			}
		}
	}
	return nil
}

func (app *IndexerApp) indexTransfersForToken(startBlock uint64, endBlock uint64, batchSize uint64, tokenAddress string, token *erc20.Erc20) error {
	decimals, err := token.Decimals(nil)
	if err != nil {
		log.Errorf("failed to get token decimals %+v", err)
		return err
	}
	name, err := token.Name(nil)
	if err != nil {
		log.Errorf("failed to get token name %+v", err)
		return err
	}
	currentBlock := startBlock
	retryCount := 20
	i := 0
	for currentBlock < endBlock {
		if i%100 == 0 {
			log.Infof("process block %d", currentBlock)
		}
		end := currentBlock + batchSize
		filterOpts := bind.FilterOpts{
			Start:   currentBlock,
			End:     &end,
			Context: context.Background(),
		}
		iter, err := token.FilterTransfer(&filterOpts, nil, nil)
		if err != nil {
			log.Errorf("failed to get transfer events for block %d, retrying: %+v", currentBlock, err)
			retryCount--
			if retryCount < 0 {
				return errors.New("GetAllTransfers failed with 0 retries")
			}
			continue
		}

		transfers := []*types.Transfer{}
		for iter.Next() {
			transferValueDecimal := utils.BigIntToFloat(iter.Event.Value, decimals)
			transfers = append(transfers,
				&types.Transfer{
					From:         iter.Event.From.String(),
					To:           iter.Event.To.String(),
					Value:        transferValueDecimal,
					BlockNumber:  iter.Event.Raw.BlockNumber,
					TxHash:       iter.Event.Raw.TxHash.String(),
					TokenAddress: tokenAddress,
					LogIndex:     uint64(iter.Event.Raw.Index),
				})
		}
		err = app.db.UpdateTokenHeight(tokenAddress, end)
		if err != nil {
			log.Warnf("failed to update token height %+v", err)
		}
		currentBlock = end

		if len(transfers) == 0 {
			continue
		}

		err = app.db.UpsertTransferBatch(transfers)
		if err != nil {
			log.Errorf("failed to upsert transfer batch %+v", err)
			return err
		}
		log.Debugf("%s: updated transfers for blocks through %d with %d transfers", name, end, len(transfers))
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}
