package engine

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/khezen/bulklog/consumer"

	"github.com/go-redis/redis"
)

func redisConvey(red *redis.Client, pipeKey string, consumers map[string]consumer.Interface) {
	startedAt, retryPeriod, retentionPeriod, err := getRedisPipe(red, pipeKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	presetRedisConvey(
		red, pipeKey,
		consumers,
		startedAt,
		retryPeriod, retentionPeriod,
	)
}

func presetRedisConvey(
	red *redis.Client, pipeKey string,
	consumers map[string]consumer.Interface,
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration) {
	documents, err := getRedisPipeDocuments(red, pipeKey)
	if err != nil {
		fmt.Printf("getRedisPipeDocuments.%s)\n", err)
		return
	}
	var (
		remainingConsumers  map[string]consumer.Interface
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
		cmders              []redis.Cmder
	)
	for {
		latestTryAt = time.Now().UTC()
		func() {
			tx = red.TxPipeline()
			defer func() {
				if err != nil {
					err = tx.Discard()
					if err != nil {
						fmt.Printf("DISCARD.%s)\n", err)
					}
				} else {
					cmders, err = tx.Exec()
					if err != nil {
						fmt.Printf("EXEC.%v.%s)\n", cmders, err)
					}
				}
			}()
			remainingConsumers, err = getRedisPipeConsumers(tx, pipeKey, consumers)
			if err != nil {
				fmt.Printf("getRedisPipeDocuments.%s)\n", err)
				return
			}
			wg = sync.WaitGroup{}
			for consumerName, cons := range remainingConsumers {
				wg.Add(1)
				go func(consumerName string, cons consumer.Interface) {
					err = cons.Digest(documents)
					if err != nil {
						fmt.Printf("Digest.%s)\n", err)
						err = nil
					} else {
						err = deleteRedisPipeConsumer(tx, pipeKey, consumerName)
						if err != nil {
							fmt.Printf("deleteRedisPipeConsumer.%s)\n", err)
							err = nil
						}
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
					err = tx.Discard()
					if err != nil {
						fmt.Printf("DISCARD.%s)\n", err)
					}
				} else {
					cmders, err := tx.Exec()
					if err != nil {
						fmt.Printf("EXEC.%v.%s)\n", cmders, err)
					}
				}
			}()
			currentTimeUnixNano = time.Now().UTC().UnixNano()
			if len(remainingConsumers) == 0 || currentTimeUnixNano > dieAtUnixNano {
				err = deleteRedisPipe(tx, pipeKey)
				if err != nil {
					fmt.Printf("deleteRedisPipe.%s)\n", err)
				} else {
					quit = true
					return
				}
			}
			iteration, err = getRedisPipeIteration(tx, pipeKey)
			if err != nil {
				fmt.Printf("getRedisPipeIteration.%s)\n", err)
				waitFor = retryPeriod - time.Since(latestTryAt)
				return
			}
			waitFor = retryPeriod*time.Duration(math.Pow(2, float64(iteration))) - time.Since(latestTryAt)
			nextTryAtUnixNano = currentTimeUnixNano + int64(waitFor)
			if nextTryAtUnixNano > dieAtUnixNano {
				err = deleteRedisPipe(tx, pipeKey)
				if err != nil {
					fmt.Printf("deleteRedisPipe.%s)\n", err)
				} else {
					quit = true
					return
				}
			}
			err = incrRedisPipeIteration(tx, pipeKey)
			if err != nil {
				fmt.Printf("incrRedisPipeIteration.%s)\n", err)
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

func redisConveyAll(red *redis.Client, pipeKeyPrefix string, consumers map[string]consumer.Interface) {
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
			fmt.Printf("KEYS.%s; Try: %d\n", err.Error(), i)
			timer = time.NewTimer(retryPeriod)
			<-timer.C
			continue
		}
		for _, pipeKey := range keys {
			go redisConvey(red, pipeKey, consumers)
		}
		success = true
		break
	}
	if !success {
		panic(fmt.Errorf("redis KEYS kept failing after %d retries", maxTries))
	}
}
