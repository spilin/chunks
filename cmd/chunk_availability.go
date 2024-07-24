package cmd

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

var FeedChunksAvailabilityCmd = &cobra.Command{
	Use:   "availability",
	Short: "Show chunk authors continuously",
	Run: func(cmd *cobra.Command, args []string) {
		blockHash, blockHeight := getLastBlockHash()
		fmt.Printf("Last block hash: %s, block height: %d\n", blockHash, blockHeight)
		for {
			block := getBlock(blockHeight)
			if block.Result.Header.Height == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			results := getChunksParallel(block.Result.Header.Hash)
			var found, notFound []int

			for shardID, res := range results {
				if res == "Found" {
					found = append(found, shardID)
				} else {
					notFound = append(notFound, shardID)
				}
			}
			// Sort the shard IDs
			sort.Ints(found)
			sort.Ints(notFound)

			blockHeight++
			fmt.Printf("Block: %v, Found: %v, Not found: %v\n", blockHeight, found, notFound)

		}
	},
}

func getChunksParallel(blockHash string) map[int]string {
	var wg sync.WaitGroup
	results := make(map[int]string)
	mu := &sync.Mutex{}
	ch := make(chan struct {
		shardID int
		res     string
	}, 6)

	for shardID := 0; shardID <= 5; shardID++ {
		wg.Add(1)
		go func(shardID int) {
			defer wg.Done()
			res := getChunk(blockHash, shardID)
			ch <- struct {
				shardID int
				res     string
			}{shardID, res}
		}(shardID)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		mu.Lock()
		results[result.shardID] = result.res
		mu.Unlock()
	}

	return results
}

func getChunk(blockHash string, shardID int) string {
	url := fmt.Sprintf("http://rpc-speedup-cache.testnet.aurora.dev/get?prev_hash=%s&shard_id=%d", blockHash, shardID)
	// fmt.Printf("Making HTTP request to %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to make HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "Not found"
	} else {

		return "Found"
	}
}
