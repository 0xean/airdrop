package snapshot

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

var snapshotGraphqlclient = graphql.NewClient("https://hub.snapshot.org/graphql", nil)
var pageSize = 1000

func getSingleProposalVotersPage(proposalId string, page int) ([]string, error) {
	var proposalVotesQuery struct {
		Votes []struct {
			Id    string `json:"id"`
			Voter string `json:"voter"`
		} `graphql:"votes (first: $pageSize, skip: $skip, where: { proposal: $proposalId }, orderBy: \"created\", orderDirection: desc)"`
		// NOTE: the above line should be a one-liner, and there should be no whitespaces
		// between `graphql:` and `"`.
	}
	variables := map[string]interface{}{
		"proposalId": proposalId,
		"pageSize":   pageSize,
		"skip":       pageSize * page,
	}
	err := snapshotGraphqlclient.Query(context.Background(), &proposalVotesQuery, variables)
	if err != nil {
		return nil, errors.Wrapf(err, "error querying votes for page %d", page)
	}
	result := make([]string, 0)
	for _, vote := range proposalVotesQuery.Votes {
		result = append(result, vote.Voter)
	}
	return result, nil
}

func getSingleProposalVoters(proposalId string) ([]string, error) {
	voters := make([]string, 0)
	page := 0
	for {
		log.Info("getting page ", page)
		pageVoters, err := getSingleProposalVotersPage(proposalId, page)
		if err != nil {
			return nil, errors.Wrapf(err, "error querying votes")
		}
		voters = append(voters, pageVoters...)
		page++
		// if this page result size is lower than the pageSize, all votes have been gathered.
		if len(pageVoters) < pageSize {
			break
		}
	}
	return voters, nil
}

func (app *SnapshotIndexerApp) GetSnapshotProposalVoters() ([]*types.SnapshotVoter, error) {
	ethChain, err := app.db.FindChain("ETH")
	if err != nil {
		return nil, errors.Wrapf(err, "error finding ETH chain")
	}

	client, err := ethclient.Dial(ethChain.RpcUrl)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to eth RPC client")
	}

	ctx := context.Background()
	block, err := client.BlockByNumber(ctx, big.NewInt(int64(ethChain.SnapshotStartBlock)))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current block from RPC client")
	}
	startBlockTime := block.Time()
	log.Infof("startBlockTime: %d", startBlockTime)

	if block, err = client.BlockByNumber(ctx, big.NewInt(int64(ethChain.SnapshotEndBlock))); err != nil {
		return nil, errors.Wrapf(err, "error getting snapshot end block")
	}
	endBlockTime := block.Time()
	log.Infof("endBlockTime: %d", endBlockTime)

	var proposalsQuery struct {
		Proposals []struct {
			Id    string `json:"id"`
			Start uint64 `json:"start"`
			End   uint64 `json:"end"`
			Title string `json:"title"`
		} `graphql:"proposals (first: 1000, where: { space: \"shapeshiftdao.eth\" }, orderBy: \"created\", orderDirection: desc)"`
		// NOTE: the above line should be a one-liner, and there should be no whitespaces
		// between `graphql:` and `"`.
	}

	// map[string]interface{}{"start": 1669759199, "end": 1676400155}
	if err = snapshotGraphqlclient.Query(context.Background(), &proposalsQuery, nil); err != nil {
		return nil, errors.Wrapf(err, "error querying proposals")
	}

	allVoters := make(map[string]*types.SnapshotVoter)
	for _, proposal := range proposalsQuery.Proposals {
		if proposal.End >= startBlockTime && proposal.End <= endBlockTime {
			// getting voters for a single proposal
			log.Info("getting voters for proposal ", proposal.Title)
			proposalVoters, err := getSingleProposalVoters(proposal.Id)
			if err != nil {
				panic(fmt.Sprintf("error getting proposal voters: %+v", err))
			}
			// appending this proposal voters
			for _, voter := range proposalVoters {
				allVoters[voter] = &types.SnapshotVoter{Address: voter}
			}
		}
	}
	result := make([]*types.SnapshotVoter, 0, len(allVoters))
	for _, voter := range allVoters {
		result = append(result, voter)
	}
	return result, nil
}
