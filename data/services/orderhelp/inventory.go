package orderhelp

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

func SetNXKeys(rdb *redis.Client, keys []string) bool {
	pipe := rdb.TxPipeline()
	setnxResults := make([]*redis.BoolCmd, len(keys))

	for i, key := range keys {
		setnxResults[i] = pipe.SetNX(context.Background(), key, "", 30*time.Second)
	}

	_, err := pipe.Exec(context.Background())
	if err != nil {
		return false
	}

	for _, res := range setnxResults {
		if !res.Val() {
			return false
		}
	}
	return true
}

func UnsetKeysInventory(rdb *redis.Client, store string, varIDs []int) {
	keys := []string{}
	for _, vid := range varIDs {
		keys = append(keys, store+"::INH::"+strconv.Itoa(vid))
	}

	pipe := rdb.TxPipeline()

	for _, key := range keys {
		pipe.Del(context.Background(), key)
	}

	_, err := pipe.Exec(context.Background())

	if err != nil {
		log.Printf("Unable to unset inventory hold keys: %v\n", err)
	}
}

func ProceedInventory(rdb *redis.Client, store string, varIDs []int) error {
	start := time.Now()

	keys := []string{}
	for _, vid := range varIDs {
		keys = append(keys, store+"::INH::"+strconv.Itoa(vid))
	}

	var retryDelay time.Duration
	for {
		if SetNXKeys(rdb, keys) {
			return nil
		}

		elapsed := time.Since(start)
		if elapsed > 40*time.Second {
			return fmt.Errorf("timed out after 40 seconds")
		}

		if elapsed <= 10*time.Second {
			retryDelay = time.Second
		} else {
			retryDelay = 3 * time.Second
		}

		time.Sleep(retryDelay)
	}
}
