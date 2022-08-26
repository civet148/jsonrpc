package main

import (
	"encoding/base64"
	"fmt"
	"github.com/civet148/jsonrpc"
	"github.com/civet148/log"
	"net/http"
	"time"
)

/*
	test JSON-RPC datagram relay to other JSON-RPC server
*/

type BlockRequest struct {
	Finality string `json:"finality"`
}

type BlockResponse struct {
		Author string `json:"author"`
		Chunks []struct {
			BalanceBurnt         string        `json:"balance_burnt"`
			ChunkHash            string        `json:"chunk_hash"`
			EncodedLength        int           `json:"encoded_length"`
			EncodedMerkleRoot    string        `json:"encoded_merkle_root"`
			GasLimit             int64         `json:"gas_limit"`
			GasUsed              int64         `json:"gas_used"`
			HeightCreated        int           `json:"height_created"`
			HeightIncluded       int           `json:"height_included"`
			OutcomeRoot          string        `json:"outcome_root"`
			OutgoingReceiptsRoot string        `json:"outgoing_receipts_root"`
			PrevBlockHash        string        `json:"prev_block_hash"`
			PrevStateRoot        string        `json:"prev_state_root"`
			RentPaid             string        `json:"rent_paid"`
			ShardId              int           `json:"shard_id"`
			Signature            string        `json:"signature"`
			TxRoot               string        `json:"tx_root"`
			ValidatorReward      string        `json:"validator_reward"`
		} `json:"chunks"`
		Header struct {
			Approvals             []string      `json:"approvals"`
			BlockMerkleRoot       string        `json:"block_merkle_root"`
			BlockOrdinal          int           `json:"block_ordinal"`
			ChallengesRoot        string        `json:"challenges_root"`
			ChunkHeadersRoot      string        `json:"chunk_headers_root"`
			ChunkMask             []bool        `json:"chunk_mask"`
			ChunkReceiptsRoot     string        `json:"chunk_receipts_root"`
			ChunkTxRoot           string        `json:"chunk_tx_root"`
			ChunksIncluded        int           `json:"chunks_included"`
			EpochId               string        `json:"epoch_id"`
			GasPrice              string        `json:"gas_price"`
			Hash                  string        `json:"hash"`
			Height                int           `json:"height"`
			LastDsFinalBlock      string        `json:"last_ds_final_block"`
			LastFinalBlock        string        `json:"last_final_block"`
			LatestProtocolVersion int           `json:"latest_protocol_version"`
			NextBpHash            string        `json:"next_bp_hash"`
			NextEpochId           string        `json:"next_epoch_id"`
			OutcomeRoot           string        `json:"outcome_root"`
			PrevHash              string        `json:"prev_hash"`
			PrevHeight            int           `json:"prev_height"`
			PrevStateRoot         string        `json:"prev_state_root"`
			RandomValue           string        `json:"random_value"`
			RentPaid              string        `json:"rent_paid"`
			Signature             string        `json:"signature"`
			Timestamp             int64         `json:"timestamp"`
			TimestampNanosec      string        `json:"timestamp_nanosec"`
			TotalSupply           string        `json:"total_supply"`
			ValidatorReward       string        `json:"validator_reward"`
		} `json:"header"`
}

func init() {
	log.SetLevel("debug")
}

func main() {
	var strUserName = "iTHwEPZ4YE54UTP4dq"
	var strPassword = "VEJLH67I9QyQx8pn0nTxmciLSWj11bhDAtNU"
	var strUrl = "ws://192.168.20.119:28082/api/v1"
	var strToken = fmt.Sprintf("%s:%s", strUserName, strPassword)
	strToken = base64.StdEncoding.EncodeToString([]byte(strToken))
	header := http.Header{}
	header.Add("Authorization", "Basic "+strToken)
	relay, err := jsonrpc.NewWebSocketClient(strUrl, header)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	defer relay.Close()

	var result BlockResponse
	err = relay.Call(&result, "block", BlockRequest{
		Finality: "final",
	})
	//err := relay.Call(&result, "block", nil)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	log.Infof("RPC relay response [%+v]", result)
	relay.Close()
	time.Sleep(3*time.Second)
}
