package cmd

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

const (
	rpcURL          = "https://rpc.testnet.near.org"
	blockHashMethod = "block"
	chunkMethod     = "chunk"
)

type RPCRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type BlockResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  BlockResult `json:"result"`
	Id      string      `json:"id"`
}

type BlockResult struct {
	Header BlockHeader `json:"header"`
}

type BlockHeader struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
}

type ChunkResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  ChunkResult `json:"result"`
	Id      string      `json:"id"`
}

type ChunkResult struct {
	Author string `json:"author"`
}

type Chunk struct {
	Block  int64  `json:"block"`
	Author string `json:"author"`
	Chunk  int    `json:"chunk"`
}

var ShowChunkAuthorsCmd = &cobra.Command{
	Use:   "show",
	Short: "Show chunk authors for the last block",
	Run: func(cmd *cobra.Command, args []string) {
		showChunkAuthors()
	},
}

var FeedChunkAuthorsCmd = &cobra.Command{
	Use:   "feed",
	Short: "Show chunk authors continuously",
	Run: func(cmd *cobra.Command, args []string) {
		blockHash, blockHeight := getLastBlockHash()
		fmt.Printf("Last block hash: %s, block height: %d\n", blockHash, blockHeight)
		for {
			// Get chunk authors for shards 0-5
			block := getBlock(blockHeight)
			if block.Result.Header.Height == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			for shardID := 0; shardID <= 5; shardID++ {
				author := getChunkAuthor(blockHeight, shardID)
				fmt.Printf("Block %d, Shard %d author: %s\n", blockHeight, shardID, author)
			}
			blockHeight++
		}
	},
}

var CollectChunkAuthorsCmd = &cobra.Command{
	Use:   "collect",
	Short: "Show chunk authors continuously",
	Run: func(cmd *cobra.Command, args []string) {
		blockHash, blockHeight := getLastBlockHash()
		db, err := connectDB()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()
		fmt.Printf("Last block hash: %s, block height: %d\n", blockHash, blockHeight)
		sleepCounter := 0
		for {
			// Get chunk authors for shards 0-5
			block := getBlock(blockHeight)
			if sleepCounter > 40 {
				log.Printf("Skipping block %d", blockHeight)
				sleepCounter = 0
				blockHeight++
				continue
			}
			if block.Result.Header.Height == 0 {
				time.Sleep(100 * time.Millisecond)
				sleepCounter++
				continue
			}
			sleepCounter = 0
			for shardID := 0; shardID <= 5; shardID++ {
				author := getChunkAuthor(blockHeight, shardID)
				fmt.Printf("Block %d, Shard %d author: %s\n", blockHeight, shardID, author)
				chunk := Chunk{
					Block:  blockHeight,
					Author: author,
					Chunk:  shardID,
				}
				_, err := db.Exec("INSERT INTO chunks (block, author, chunk) VALUES ($1, $2, $3)",
					chunk.Block, chunk.Author, chunk.Chunk)
				if err != nil {
					log.Printf("Failed to insert chunk data: %v", err)
				}
			}
			blockHeight++
		}
	},
}

func showChunkAuthors() {
	blockHash, blockHeight := getLastBlockHash()
	fmt.Printf("Last block hash: %s, block height: %d\n", blockHash, blockHeight)

	// Get chunk authors for shards 0-5
	for shardID := 0; shardID <= 5; shardID++ {
		author := getChunkAuthor(blockHeight, shardID)
		fmt.Printf("Shard %d author: %s\n", shardID, author)
	}
}

func getLastBlockHash() (string, int64) {
	params := map[string]string{
		"finality": "final",
	}
	reqBody, _ := json.Marshal(RPCRequest{
		Jsonrpc: "2.0",
		Id:      "dontcare",
		Method:  blockHashMethod,
		Params:  params,
	})

	respBody := makeRPCRequest(reqBody)
	var blockResp BlockResponse
	json.Unmarshal(respBody, &blockResp)

	return blockResp.Result.Header.Hash, blockResp.Result.Header.Height
}

func getBlock(blockHeight int64) BlockResponse {
	params := map[string]interface{}{
		"block_id": blockHeight,
	}
	reqBody, _ := json.Marshal(RPCRequest{
		Jsonrpc: "2.0",
		Id:      "dontcare",
		Method:  blockHashMethod,
		Params:  params,
	})

	respBody := makeRPCRequest(reqBody)
	var blockResp BlockResponse
	json.Unmarshal(respBody, &blockResp)

	return blockResp
}

func getChunkAuthor(blockHeight int64, shardID int) string {
	params := map[string]interface{}{
		"block_id": blockHeight,
		"shard_id": shardID,
	}
	reqBody, _ := json.Marshal(RPCRequest{
		Jsonrpc: "2.0",
		Id:      "dontcare",
		Method:  chunkMethod,
		Params:  params,
	})

	respBody := makeRPCRequest(reqBody)
	var chunkResp ChunkResponse
	json.Unmarshal(respBody, &chunkResp)

	return chunkResp.Result.Author
}

func makeRPCRequest(reqBody []byte) []byte {
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("Failed to make RPC request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	// fmt.Printf("------------------------------------------------\n")
	// fmt.Printf("Response: %s\n", body)
	return body
}

func connectDB() (*sql.DB, error) {
	connStr := "postgres://spilin:@localhost:5432/copy?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}
