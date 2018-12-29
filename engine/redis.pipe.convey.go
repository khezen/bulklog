package engine

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/khezen/bulklog/consumer"

	"github.com/go-redis/redis"
)

func redisConvey(red *redis.Client, pipeKey string) {
	startedAt, retryPeriod, retentionPeriod, err := getRedisPipe(red, pipeKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	presetRedisConvey(
		red, pipeKey,
		startedAt,
		retryPeriod, retentionPeriod,
	)
}

func presetRedisConvey(
	red *redis.Client, pipeKey string,
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration) {
	documents, err := getRedisDocuments(red, pipeKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	var (
		consumers           map[string]consumer.Interface
		dieAt               = startedAt.Add(retentionPeriod)
		dieAtUnixNano       = dieAt.UnixNano()
		currentTimeUnixNano int64
		nextTryAtUnixNano   int64
		iteration           int
		latestTryAt         time.Time
		waitFor             time.Duration
		timer               *time.Timer
		tx                  redis.Pipeliner
		wg                  sync.WaitGroup
		quit                bool
	)
	for {
		latestTryAt = time.Now().UTC()
		func() {
			tx = red.TxPipeline()
			defer func() {
				if err != nil {
					tx.Discard()
				} else {
					tx.Exec()
				}
			}()
			consumers, err = getRedisConsumers(tx, pipeKey)
			if err != nil {
				fmt.Println(err)
				return
			}
			wg = sync.WaitGroup{}
			for consumerName, cons := range consumers {
				wg.Add(1)
				go func(consumerName string, cons consumer.Interface) {
					err = cons.Digest(documents)
					if err != nil {
						fmt.Println(err)
						err = nil
					} else {
						err = delRedisConsumer(tx, pipeKey, consumerName)
						if err != nil {
							fmt.Println(err)
							err = nil
						}
						delete(consumers, consumerName)
					}
					wg.Done()
				}(consumerName, cons)
			}
			wg.Wait()
		}()
		func() {
			tx = red.TxPipeline()
			defer func() {
				if err != nil {
					tx.Discard()
				} else {
					tx.Exec()
				}
			}()
			currentTimeUnixNano = time.Now().UTC().UnixNano()
			if len(consumers) == 0 || currentTimeUnixNano > dieAtUnixNano {
				err = deleteRedisPipe(tx, pipeKey)
				if err != nil {
					fmt.Println(err)
				} else {
					quit = true
					return
				}
			}
			iteration, err = getRedisIteration(tx, pipeKey)
			if err != nil {
				fmt.Println(err)
				waitFor = retryPeriod - time.Since(latestTryAt)
				return
			}
			waitFor = retryPeriod*time.Duration(math.Pow(2, float64(iteration))) - time.Since(latestTryAt)
			nextTryAtUnixNano = currentTimeUnixNano + int64(waitFor)
			if nextTryAtUnixNano > dieAtUnixNano {
				err = deleteRedisPipe(tx, pipeKey)
				if err != nil {
					fmt.Println(err)
				} else {
					quit = true
					return
				}
			}
			err = incrRedisIteration(tx, pipeKey)
			if err != nil {
				fmt.Println(err)
				return
			}
		}()
		if quit {
			return
		}
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}

func redisConveyAll(red *redis.Client, pipeKeyPrefix string) {
	var (
		pattern     = fmt.Sprintf(`%s\..{36}$`, pipeKeyPrefix)
		maxTries    = 20
		retryPeriod = 10 * time.Second
		timer       *time.Timer
		success     bool
	)
	for i := 0; i < maxTries; i++ {
		keys, err := red.Keys(pattern).Result()
		if err != nil {
			fmt.Printf("redis connection failed. Try: %d\n", i)
			timer = time.NewTimer(retryPeriod)
			<-timer.C
			continue
		}
		for _, pipeKey := range keys {
			go redisConvey(red, pipeKey)
		}
		success = true
		break
	}
	if !success {
		panic(fmt.Errorf("redis connection kept failing after %d retries", maxTries))
	}
}
