package indexer

import (
	"encoding/json"
	"fmt"
	"github.com/ArkeoNetwork/airdrop/pkg/db"
	"github.com/ArkeoNetwork/airdrop/pkg/types"
	"github.com/ArkeoNetwork/airdrop/pkg/utils"
	arkutils "github.com/ArkeoNetwork/common/utils"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	// abcitypes "github.com/tendermint/tendermint/abci/types"
	"io"
	"math/big"
	// "net"
	"net/http"
	"net/url"
	// "net/rpc"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	// coretypes "github.com/tendermint/tendermint/rpc/core/types"
	// tmtypes "github.com/tendermint/tendermint/types"
	"os"
	"strconv"
	"strings"
	"time"
)

type CosmosIndexerParams struct {
	DB    arkutils.DBConfig
	Chain string
}

type CosmosIndexer struct {
	db          *db.AirdropDB
	tm          *tmhttp.HTTP
	lcd         *resty.Client
	chain       *types.Chain
	rpc         string
	startHeight int64
	endHeight   int64
}

// a page of delegator/validator pairs with amounts for starting balances
type DelegationPage struct {
	DelegationResponses []DelegationResponse `json:"delegation_responses"`
}

type Delegation struct {
	DelegatorAddress string `json:"delegator_address"`
	ValidatorAddress string `json:"validator_address"`
	Shares           string `json:"shares"`
}

type DelegationResponse struct {
	Delegation Delegation `json:"delegation"`
	Balance    struct {
		Denom  string `json:"denom"`
		Amount int64  `json:"amount,string"`
	} `json:"balance"`
}

type Liquidity struct {
	Position struct {
		Address   string `json:"address"`
		Liquidity string `json:"liquidity"`
		PoolId    string `json:"pool_id"`
	} `json:"position"`
}

type ImportedDelegation struct {
	AppState struct {
		Staking struct {
			Delegations []Delegation `json:"delegations"`
		} `json:"staking"`
		ConcentratedLiquidity struct {
			PositionData []Liquidity `json:"position_data"`
		} `json:"concentratedliquidity"`
	} `json:"app_state"`
}

type TxSearchResponse struct {
	Result Result `json:"result"`
}

type Result struct {
	Txs        []*Tx  `json:"txs"`
	TotalCount string `json:"total_count"`
}

type Tx struct {
	Hash     string   `json:"hash"`
	Height   string   `json:"height"`
	TxResult TxResult `json:"tx_result"`
	Tx       string   `json:"tx"`
}

type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index bool   `json:"index"`
}

type Event struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

type TxResult struct {
	Code      int     `json:"code"`
	Log       string  `json:"log"`
	Info      string  `json:"info"`
	GasWanted string  `json:"gas_wanted"`
	GasUsed   string  `json:"gas_used"`
	Events    []Event `json:"events"`
}

func NewCosmosIndexer(params CosmosIndexerParams) (*CosmosIndexer, error) {
	d, err := db.New(params.DB)
	if err != nil {
		return nil, errors.Wrapf(err, "error connecting to db")
	}
	chain, err := d.FindChain(params.Chain)
	if err != nil {
		return nil, errors.Wrapf(err, "error finding chain %s", params.Chain)
	}

	log.Infof("connecting to tendermint node at %s", chain.RpcUrl)
	rpc := chain.RpcUrl
	lcd := resty.New().SetTimeout(10 * time.Second).SetBaseURL(chain.LcdUrl)

	return &CosmosIndexer{db: d, rpc: rpc, lcd: lcd, chain: chain, startHeight: int64(chain.SnapshotStartBlock), endHeight: int64(chain.SnapshotEndBlock)}, nil
}

func (c *CosmosIndexer) IndexDelegationsFromStateExport(stateExportFile, chain string, height int64) error {
	start := time.Now()
	raw, err := os.ReadFile(stateExportFile)
	if err != nil {
		return errors.Wrapf(err, "error reading file %s", stateExportFile)
	}
	log.Infof("read file %s in %.3f seconds", stateExportFile, time.Since(start).Seconds())
	start = time.Now()
	imported := ImportedDelegation{}
	if err = json.Unmarshal(raw, &imported); err != nil {
		return errors.Wrapf(err, "error unmarshalling file %s", stateExportFile)
	}
	log.Infof("unmarshalled delegations %s in %.3f seconds", stateExportFile, time.Since(start).Seconds())

	events := make([]*types.CosmosStakingEvent, 0, len(imported.AppState.Staking.Delegations))

	start = time.Now()
	for _, d := range imported.AppState.Staking.Delegations {
		value, err := parseShares(d.Shares, c.chain.Decimals)
		if err != nil {
			return errors.Wrapf(err, "%s delegation to %s error parsing shares %s", d.DelegatorAddress, d.ValidatorAddress, d.Shares)
		}
		if value <= 0 {
			log.Warnf("%s delegation to %s with value %f. string shares: %s", d.DelegatorAddress, d.ValidatorAddress, value, d.Shares)
		}
		event := &types.CosmosStakingEvent{
			Chain:       c.chain.Name,
			EventType:   "initial",
			Delegator:   d.DelegatorAddress,
			Validator:   d.ValidatorAddress,
			Value:       value,
			BlockNumber: uint64(height),
			TxHash:      "00000000000000000000000000000000",
			EventIndex:  0,
		}
		events = append(events, event)
	}
	log.Infof("created %d staking events in %.3f seconds", len(events), time.Since(start).Seconds())
	start = time.Now()
	if err = c.db.InsertStakingEvents(events); err != nil {
		return errors.Wrapf(err, "error inserting staking events")
	}
	log.Infof("inserted %d staking events in %.3f seconds", len(events), time.Since(start).Seconds())
	return nil
}

func (c *CosmosIndexer) IndexLiquidityFromStateExport(stateExportFile, chain string, height int64) error {
	start := time.Now()
	raw, err := os.ReadFile(stateExportFile)
	if err != nil {
		return errors.Wrapf(err, "error reading file %s", stateExportFile)
	}
	log.Infof("read file %s in %.3f seconds", stateExportFile, time.Since(start).Seconds())
	start = time.Now()
	imported := ImportedDelegation{}
	if err = json.Unmarshal(raw, &imported); err != nil {
		return errors.Wrapf(err, "error unmarshalling file %s", stateExportFile)
	}
	log.Infof("unmarshalled delegations %s in %.3f seconds", stateExportFile, time.Since(start).Seconds())
	log.Infof("Number of entries in PositionData: %d", len(imported.AppState.ConcentratedLiquidity.PositionData))

	events := make([]*types.OsmoLP, 0, len(imported.AppState.ConcentratedLiquidity.PositionData))

	start = time.Now()
	for _, d := range imported.AppState.ConcentratedLiquidity.PositionData {
		log.Infof("processing liquidity for %s", d.Position.Address)
		value, err := parseBigFloat(d.Position.Liquidity)
		if err != nil {
			return errors.Wrapf(err, "%s liquidity to %s error parsing shares %s", d.Position.Address, d.Position.Address, d.Position.Liquidity)
		}
		var eventType string

		zero := new(big.Float).SetInt64(0)
		cmp := value.Cmp(zero)
		if cmp > 0 {
			eventType = "create"
		} else {
			eventType = "withdraw"
		}
		poolId, err := strconv.ParseInt(d.Position.PoolId, 10, 64)
		if err != nil {
			log.Errorf("Error converting string to int64: %v", err)
			return errors.Wrapf(err, "Error converting pool id string")
		}
		event := &types.OsmoLP{
			Type:        eventType,
			LpAmount:    value.Text('f', -1),
			BlockNumber: height,
			Account:     d.Position.Address,
			PoolId:      poolId,
			TxHash:      "00000000000000000000000000000000",
		}
		events = append(events, event)
	}
	log.Infof("created %d liquidity events in %.3f seconds", len(events), time.Since(start).Seconds())
	start = time.Now()
	if err = c.db.InsertOsmoLP(events); err != nil {
		return errors.Wrapf(err, "error inserting liquidity events")
	}
	log.Infof("inserted %d liquidity events in %.3f seconds", len(events), time.Since(start).Seconds())
	return nil
}

func parseShares(s string, decimals uint8) (float64, error) {
	if !strings.Contains(s, ".") {
		return -1, fmt.Errorf("shares %s does not contain a decimal", s)
	}
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return -1, fmt.Errorf("shares %s has more than one decimal", s)
	}
	ishares, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return -1, errors.Wrapf(err, "error parsing shares %d", ishares)
	}

	return utils.FromBaseUnits(ishares, decimals), nil
}

func (c *CosmosIndexer) IndexCosmosDelegators() error {
	startHeight := int64(c.chain.SnapshotStartBlock)
	endHeight := int64(c.chain.SnapshotEndBlock)

	latest, err := c.db.FindLatestIndexedCosmosStakingBlock(c.chain.Name)
	if err != nil {
		return errors.Wrapf(err, "error finding latest indexed block")
	}
	if latest > startHeight {
		log.Infof("found latest indexed block %d, starting at %d", latest, latest-1)
		startHeight = latest - 1
	}

	// for i := startHeight; i <= endHeight; i++ {
	// 	if err := c.indexCosmosDelegations(i); err != nil {
	// 		log.Errorf("error indexing delegations at height %d: %+v", i, err)
	// 		break
	// 	}
	// }

	const retryInterval = 5

	for i := startHeight; i <= endHeight; i++ {
		for {
			if err := c.indexCosmosDelegations(i); err != nil {
				log.Errorf("error indexing delegations at height %d: %+v", i, err)
	
				// Log and wait for 5 seconds before retrying
				log.Warnf("Retrying height %d in %d seconds...", i, retryInterval)
				time.Sleep(time.Duration(retryInterval) * time.Second)
				continue // Retry indefinitely until the call succeeds
			}
	
			// If no error, move to the next height
			log.Debugf("Successfully indexed height %d", i)
			break
		}
	}
	return nil
}

func shouldStoreLPTx(txResults *TxResult) bool {
	for _, evt := range txResults.Events {
		switch evt.Type {
		case "withdraw_position", "create_position":
			return true
		default:
		}
	}
	return false
}

func shouldStoreTx(txResults *TxResult) bool {
	for _, evt := range txResults.Events {
		switch evt.Type {
		case "delegate":
			return true
		case "unbond":
			return true
		case "redelegate":
			return true
		default:
		}
	}
	return false
}

func getAttribute(attributes []EventAttribute, attributeName string) string {
	for _, attr := range attributes {
		if attr.Key == attributeName {
			return attr.Value
		}
	}
	return ""
}

func (c *CosmosIndexer) handleOsmoLPTx(height int64, txHash string, txResult *TxResult) error {
	log := log.WithField("height", fmt.Sprintf("%d", height))
	log.Debugf("handling osmo lp tx %s", txHash)
	liquidityEvents := make([]*types.OsmoLP, 0, len(txResult.Events))

	for _, evt := range txResult.Events {
		switch evt.Type {
		case "withdraw_position", "create_position":
			poolIdStr := getAttribute(evt.Attributes, "pool_id")
			if poolIdStr != "" {
				log.Debugf("Pool ID:", poolIdStr)
			} else {
				log.Errorf("Pool ID field not found")
			}
			poolId, err := strconv.ParseInt(poolIdStr, 10, 64)
			if err != nil {
				log.Errorf("Error converting string to int64: %v", err)
				return errors.Wrapf(err, "Error converting pool id string")

			}
			sender := getAttribute(evt.Attributes, "sender")
			if sender != "" {
				log.Debugf("Sender:", sender)
			} else {
				log.Errorf("Sender field not found")
			}

			liquidityStr := getAttribute(evt.Attributes, "liquidity")
			if liquidityStr != "" {
				log.Debugf("Liquidity:", liquidityStr)
			} else {
				log.Errorf("Liquidity field not found")
				return errors.Wrapf(err, "Liquidity field not found")
			}
			liquidity, err := parseBigFloat(liquidityStr)
			if err != nil {
				log.Errorf("error parsing amount %s: %+v", liquidityStr, err)
				return errors.Wrapf(err, "error parsing amount %s", liquidityStr)
			}
			withdrawEvt := types.OsmoLP{
				Type:        strings.Replace(evt.Type, "_position", "", 1),
				LpAmount:    liquidity.Text('f', -1),
				Account:     sender,
				BlockNumber: int64(height),
				TxHash:      txHash,
				PoolId:      poolId,
			}
			liquidityEvents = append(liquidityEvents, &withdrawEvt)
		}
	}
	log.Infof("inserting %d liquidity events", len(liquidityEvents))
	if len(liquidityEvents) == 0 {
		return errors.New("no liquidity events to insert")
	}
	return c.db.InsertOsmoLP(liquidityEvents)
}

func (c *CosmosIndexer) handleStakingTx(height int64, txHash string, txResult *TxResult) error {
	log := log.WithField("height", fmt.Sprintf("%d", height))
	evtsSequenced := make([]Event, len(txResult.Events))
	evtsSeq := int64(0)
	evtsIndexMap := make(map[int64]int64, 1024)
	for i, evt := range txResult.Events {
		switch evt.Type {
		case "delegate":
			evtsSequenced[evtsSeq] = evt
			evtsIndexMap[evtsSeq] = int64(i)
			evtsSeq++
		case "redelegate":
			evtsSequenced[evtsSeq] = evt
			evtsIndexMap[evtsSeq] = int64(i)
			evtsSeq++
		case "unbond":
			evtsSequenced[evtsSeq] = evt
			evtsIndexMap[evtsSeq] = int64(i)
			evtsSeq++
		case "message":
			m := make(map[string]string, len(evt.Attributes))
			for _, attr := range evt.Attributes {
				m[string(attr.Key)] = string(attr.Value)
			}
			if module, ok := m["module"]; ok && (module == "staking") {
				if delegator, ok := m["sender"]; ok {
					log.Debugf("adding delegate event delegator %s", delegator)
					evtsSequenced[evtsSeq] = evt
					evtsSeq++
				}
			}
		}
	}
	log.Debugf("evtsSeq %v", evtsSequenced)
	stakingEvents := make([]*types.CosmosStakingEvent, 0, len(evtsSequenced))
	evtsSequenced = evtsSequenced[:evtsSeq]
	var stakingEvt *stakingEventWrapper
	log.Debugf("evtsSequenced %d", len(evtsSequenced))
	for i, evt := range evtsSequenced {
		log.Debugf("EVT: %s, Index: %d", evt.Type, i)
		m := attributesToMap(evt.Attributes)
		if evt.Type != "message" {
			// staking event itself: delegate,unbond,redelegate
			_, amount, err := parseAmount(m["amount"], c.chain.Decimals)
			if err != nil {
				log.Errorf("error parsing amount %s: %+v", m["amount"], err)
				return errors.Wrapf(err, "error parsing amount %s", m["amount"])
			}
			var srcValidator, destValidator, delegator string
			var delegatorExists bool
			switch evt.Type {
			case "delegate":
				destValidator = m["validator"]
				delegator, delegatorExists = m["delegator"]
			case "redelegate":
				destValidator = m["destination_validator"]
				srcValidator = m["source_validator"]
				delegator, delegatorExists = m["delegator"]
			case "unbond":
				srcValidator = m["validator"]
				amount = -amount
			}

			validator := destValidator
			if evt.Type == "unbond" {
				validator = srcValidator
			}

			evtIndex, ok := evtsIndexMap[int64(i)]
			if !ok {
				log.Warnf("no event index found for %d", i)
				evtIndex = 0
			}

			if stakingEvt != nil { // message came first
				log.Debugf("staking event second %s", delegator)
				stakingEvt.srcValidator = srcValidator
				stakingEvt.CosmosStakingEvent.EventType = evt.Type
				stakingEvt.CosmosStakingEvent.EventIndex = evtIndex
				stakingEvt.CosmosStakingEvent.Value = amount
				stakingEvt.CosmosStakingEvent.Validator = validator
				if !delegatorExists {
					stakingEvents = append(stakingEvents, &stakingEvt.CosmosStakingEvent)
				}
				if evt.Type == "redelegate" {
					stakingEvt.EventType = "delegate"
					unbondEvt := types.CosmosStakingEvent{
						EventType:   "unbond",
						EventIndex:  0, // using zero here for extra row to maintain uniqueness
						Delegator:   stakingEvt.Delegator,
						Validator:   stakingEvt.srcValidator,
						Chain:       c.chain.Name,
						Value:       -stakingEvt.Value,
						BlockNumber: uint64(height),
						TxHash:      txHash,
					}
					stakingEvents = append(stakingEvents, &unbondEvt)
				}
				if !delegatorExists {
					stakingEvt = nil
				}
			} else { // staking event came first
				log.Debugf("staking event first %s", delegator)
				stakingEvt = &stakingEventWrapper{
					CosmosStakingEvent: types.CosmosStakingEvent{
						EventType:   evt.Type,
						EventIndex:  evtIndex,
						Validator:   validator,
						Delegator:   delegator,
						Chain:       c.chain.Name,
						Value:       amount,
						BlockNumber: uint64(height),
						TxHash:      txHash,
					},
					srcValidator: srcValidator,
				}

				var delegatorInNextEvent bool
				if i+1 < len(evtsSequenced) {
					delegatorInNextEvent = evtsSequenced[i+1].Type == "message"
				} else {
					// Set a default value or handle the out-of-bounds case
					delegatorInNextEvent = false
				}

				if stakingEvt.Delegator == "" && !delegatorInNextEvent { // Authz but no delegator on event (will be last coin spent event)
					log.Debugf("Evt %v", evt)
					for _, event := range txResult.Events {
						attr := attributesToMap(event.Attributes)
						log.Debugf("Event %v", event)
						authzIndex := attr["authz_msg_index"]
						if event.Type == "coin_spent" && authzIndex == m["authz_msg_index"] {
							stakingEvt.Delegator = attr["spender"]
							delegatorExists = true
						}
					}
				}
			}

			if delegatorExists {
				stakingEvents = append(stakingEvents, &stakingEvt.CosmosStakingEvent)
				stakingEvt = nil
			}
		} else {
			log.Debugf("stakingEvt %v", stakingEvt)
			if stakingEvt == nil { // message event first
				log.Debugf("message event first")
				stakingEvt = &stakingEventWrapper{
					CosmosStakingEvent: types.CosmosStakingEvent{
						Delegator:   m["sender"],
						Chain:       c.chain.Name,
						BlockNumber: uint64(height),
						TxHash:      txHash,
					},
				}
			} else {
				log.Debugf("message event second")
				if stakingEvt.Delegator == "" {
					stakingEvt.Delegator = m["sender"]
					stakingEvents = append(stakingEvents, &stakingEvt.CosmosStakingEvent)
				}
				if stakingEvt.EventType == "redelegate" {
					log.Debugf("LENGTH: %d", len(stakingEvents))
					stakingEvents[0].EventType = "delegate"

					unbondEvt := types.CosmosStakingEvent{
						EventType:   "unbond",
						EventIndex:  0, // using zero here for extra row to maintain uniqueness
						Delegator:   stakingEvt.Delegator,
						Validator:   stakingEvt.srcValidator,
						Chain:       c.chain.Name,
						Value:       -stakingEvt.Value,
						BlockNumber: uint64(height),
						TxHash:      txHash,
					}
					stakingEvents = append(stakingEvents, &unbondEvt)
				}
				stakingEvt = nil
			}
		}
	}
	// log.Infof("stakingEvt %v", stakingEvt)
	log.Infof("inserting %d staking events", len(stakingEvents))
	if len(stakingEvents) == 0 {
		return errors.New("no staking events to insert")
	}
	return c.db.InsertStakingEvents(stakingEvents)
}

type stakingEventWrapper struct {
	types.CosmosStakingEvent
	srcValidator string
}

func attributesToMap(attributes []EventAttribute) map[string]string {
	m := make(map[string]string, len(attributes))
	for _, a := range attributes {
		m[string(a.Key)] = string(a.Value)
	}
	return m
}

func (c *CosmosIndexer) indexCosmosDelegations(height int64) error {
	log := log.WithField("height", fmt.Sprintf("%d", height))
	var (
		txSearchResults []*Tx
		txSearchErr     error
	)

	page := 1
	perPage := 100
	txSearchResults = make([]*Tx, 0, 128)

	for {
		query := url.QueryEscape(fmt.Sprintf("\"tx.height=%d\"", height))
		query = fmt.Sprintf("%s&page=%d&limit=%d", query, page, perPage)
		baseURL := c.rpc
		url := fmt.Sprintf("%s/tx_search?query=%s", baseURL, query)

		log.Debugf("Requesting %s", url)

		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to make the request: %v", err)
			txSearchErr = errors.Wrapf(err, "Failed to make the request: %v", err)
			break
		}
		defer resp.Body.Close()

		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			log.Errorf("Failed to get a successful response: %s", resp.Status)
			txSearchErr = errors.New("Failed to get a successful response")
			break
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Failed to read the response body: %v", err)
			txSearchErr = errors.Wrapf(err, "Failed to read the response body: %v", err)
			break
		}

		searchResults := TxSearchResponse{}
		if err = json.Unmarshal(body, &searchResults); err != nil {
			log.Errorf("Failed to unmarshal transaction body")
			txSearchErr = errors.Wrapf(err, "error unmarshalling response")
			break
		}

		totalCount, err := strconv.Atoi(searchResults.Result.TotalCount)
		if err != nil {
			fmt.Println("Error converting string to int:", err)
			txSearchErr = errors.Wrapf(err, "error converting string to int")
			break
		}

		if totalCount == 0 {
			log.Debugf("height %d no txs found", height)
			break
		}

		txSearchResults = append(txSearchResults, searchResults.Result.Txs...)
		if len(txSearchResults) == totalCount {
			log.Debugf("height %d break tx search loop with %d gathered. %d in page %d totalCount %s", height, len(txSearchResults), len(searchResults.Result.Txs), page, searchResults.Result.TotalCount)
			break
		}
		page++
	}

	if txSearchErr != nil {
		log.Printf("Tx Search Err %d", txSearchErr)
		return errors.Wrapf(txSearchErr, "error searching txs block %d", height)
	}

	for _, t := range txSearchResults {
		if shouldStoreTx(&t.TxResult) {
			if err := c.handleStakingTx(height, t.Hash, &t.TxResult); err != nil {
				log.Errorf("error handling staking tx %s: %+v", t.Hash, err)
				return errors.Wrapf(err, "error handling staking tx %s", t.Hash)
			}
		}
		// Handle LP txs for OSMO
		if c.chain.Name == "OSMO" && shouldStoreLPTx(&t.TxResult) {
			if err := c.handleOsmoLPTx(height, t.Hash, &t.TxResult); err != nil {
				log.Errorf("error handling osmo lp tx %s: %+v", t.Hash, err)
				return errors.Wrapf(err, "error handling osmo lp tx %s", t.Hash)
			}
		}
	}
	return nil
}

func parseAmount(in string, decimals uint8) (asset string, amount float64, err error) {
	amt := make([]byte, len(in))
	var i int
	for i = 0; i < len(in); i++ {
		if in[i] < '0' || in[i] > '9' {
			break
		}
		amt[i] = in[i]
	}
	amt = amt[:i]
	asset = in[i:]
	samt := string(amt)
	log.Debugf("%s %s", samt, asset)

	iamt, ok := new(big.Int).SetString(samt, 10)
	if !ok {
		log.Errorf("unable to convert amount %s to big int", samt)
		return
	}

	amount = utils.BigIntToFloat(iamt, uint8(decimals))
	return
}

func parseBigFloat(numberStr string) (*big.Float, error) {
	// Create a new big.Float and set its value from the string
	f := new(big.Float)
	_, _, err := f.Parse(numberStr, 10)
	if err != nil {
		return nil, fmt.Errorf("error parsing string to big.Float: %w", err)
	}
	return f, nil
}
