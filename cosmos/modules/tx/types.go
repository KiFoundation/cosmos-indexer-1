package tx

import (
	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	cosmTx "github.com/cosmos/cosmos-sdk/types/tx"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GetTxsEventResponseWrapper struct {
	CosmosGetTxsEventResponse *cosmTx.GetTxsEventResponse
	Height                    int64
}

type GetBlockByHeightResponse struct {
	BlockID BlockID       `json:"block_id"`
	Block   BlockResponse `json:"block"`
}

type BlockResponse struct {
	BlockData   BlockData   `json:"data"`
	BlockHeader BlockHeader `json:"header"`
}

type BlockID struct {
	Hash string `json:"hash"`
}

type BlockData struct {
	Txs []string `json:"txs"`
}

type BlockHeader struct {
	Height string `json:"height"`
}

type GetTxByBlockHeightResponse struct {
	Txs         []IndexerTx `json:"txs"`
	TxResponses []Response  `json:"tx_responses"`
	Pagination  Pagination  `json:"pagination"`
}

type IndexerTx struct {
	Body     Body `json:"body"`
	AuthInfo cosmTx.AuthInfo
	Signers  []sdk.AccAddress
}

type Response struct {
	TxHash    string       `json:"txhash"`
	Height    string       `json:"height"`
	TimeStamp string       `json:"timestamp"`
	Code      uint32       `json:"code"`
	RawLog    string       `json:"raw_log"`
	Log       []LogMessage `json:"logs"`
}

// TxLogMessage:
// Cosmos blockchains return Transactions with an array of "logs" e.g.
//
// "logs": [
//
//	{
//		"msg_index": 0,
//		"events": [
//		  {
//			"type": "coin_received",
//			"attributes": [
//			  {
//				"key": "receiver",
//				"value": "juno128taw6wkhfq29u83lmh5qyfv8nff6h0w577vsy"
//			  }, ...
//			]
//		  } ...
//
// The individual log always has a msg_index corresponding to the Message from the Transaction.
// But the events are specific to each Message type, for example MsgSend might be different from
// any other message type.
//
// This struct just parses the KNOWN fields and leaves the other fields as raw JSON.
// More specific type parsers for each message type can parse those fields if they choose to.
type LogMessage struct {
	MessageIndex int               `json:"msg_index"`
	Events       []LogMessageEvent `json:"events"`
}

type Attribute struct {
	Key   string
	Value string
}

type LogMessageEvent struct {
	Type       string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Body struct {
	Messages []sdk.Msg `json:"messages"`
}

type AuthInfo struct {
	TxFee         Fee          `json:"fee"`
	TxSignerInfos []SignerInfo `json:"signer_infos"` // this is used in REST but not RPC parsers
}

type Fee struct {
	TxFeeAmount []FeeAmount `json:"amount"`
	GasLimit    string      `json:"gas_limit"`
}

type FeeAmount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type SignerInfo struct {
	PublicKey PublicKey `json:"public_key"`
}

type PublicKey struct {
	Type string `json:"@type"`
	Key  string `json:"key"`
}

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

// In the json, TX data is split into 2 arrays, used to merge the full dataset
type MergedTx struct {
	Tx         IndexerTx
	TxResponse Response
}

type GetLatestBlockResponse struct {
	BlockID BlockID       `json:"block_id"`
	Block   BlockResponse `json:"block"`
}

type Message struct {
	Type string `json:"@type"`
}

func (sf *Message) GetType() string {
	return sf.Type
}

// CosmosMessage represents a Cosmos blockchain Message (part of a transaction).
// CosmUnmarshal() unmarshals the specific cosmos message type (e.g. MsgSend).
// First arg must always be the message type itself, as this won't be parsed in CosmUnmarshal.
type CosmosMessage interface {
	HandleMsg(string, sdk.Msg, *LogMessage) error
	ParseRelevantData() []parsingTypes.MessageRelevantInformation
	GetType() string
	String() string
}
