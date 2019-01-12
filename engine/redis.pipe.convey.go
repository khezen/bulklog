package engine

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/bulklog/bulklog/consumer"
)

func redisConvey(red *redis.Pool, pipeKey string, consumers map[string]consumer.Interface) {
	startedAt, retryPeriod, retentionPeriod, err := getRedisPipe(red, pipeKey)
	if err == errRedisPipeNotFound {
		err = deleteRedisPipe(red, pipeKey)
		if err != nil {
			fmt.Printf("deleteRedisPipe.%s)\n", err)
		}
		return
	}
	if err != nil {
		fmt.Printf("getRedisPipe.%s)\n", err)
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
	red *redis.Pool, pipeKey string,
	consumers map[string]consumer.Interface,
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration) {
	var (
		dieAt               = startedAt.Add(retentionPeriod)
		dieAtUnixNano       = dieAt.UnixNano()
		currentTimeUnixNano = time.Now().UTC().UnixNano()
		err                 error
	)
	if currentTimeUnixNano > dieAtUnixNano {
		err = deleteRedisPipe(red, pipeKey)
		if err != nil {
			fmt.Printf("deleteRedisPipe.%s)\n", err)
		} else {
			return
		}
	}
	documents, err := getRedisPipeDocuments(red, pipeKey)
	if err != nil {
		fmt.Printf("getRedisPipeDocuments.%s)\n", err)
		return
	}
	if len(documents) == 0 {
		err = deleteRedisPipe(red, pipeKey)
		if err != nil {
			fmt.Printf("deleteRedisPipe.%s)\n", err)
		}
		return
	}
	var (
		remainingConsumers map[string]consumer.Interface
		nextTryAtUnixNano  int64
		iteration          int
		latestTryAt        time.Time
		waitFor            time.Duration
		timer              *time.Timer
		wg                 sync.WaitGroup
	)
	for {
		latestTryAt = time.Now().UTC()
		remainingConsumers, err = getRedisPipeConsumers(red, pipeKey, consumers)
		if err != nil {
			fmt.Printf("getRedisPipeConsumers.%s)\n", err)
			return
		}
		if len(remainingConsumers) == 0 {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				fmt.Printf("deleteRedisPipe.%s)\n", err)
			}
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
					err = deleteRedisPipeConsumer(red, pipeKey, consumerName)
					if err != nil {
						fmt.Printf("deleteRedisPipeConsumer.%s)\n", err)
					}
				}
				wg.Done()
			}(consumerName, cons)
		}
		wg.Wait()
		currentTimeUnixNano = time.Now().UTC().UnixNano()
		if len(remainingConsumers) == 0 || currentTimeUnixNano > dieAtUnixNano {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				fmt.Printf("deleteRedisPipe.%s)\n", err)
			} else {
				return
			}
		}
		iteration, err = getRedisPipeIteration(red, pipeKey)
		if err != nil {
			fmt.Printf("getRedisPipeIteration.%s)\n", err)
			waitFor = retryPeriod - time.Since(latestTryAt)
			continue
		}
		waitFor = retryPeriod*time.Duration(math.Pow(2, float64(iteration))) - time.Since(latestTryAt)
		nextTryAtUnixNano = currentTimeUnixNano + int64(waitFor)
		if nextTryAtUnixNano > dieAtUnixNano {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				fmt.Printf("deleteRedisPipe.%s)\n", err)
			} else {
				return
			}
		}
		err = incrRedisPipeIteration(red, pipeKey)
		if err != nil {
			fmt.Printf("incrRedisPipeIteration.%s)\n", err)
			return
		}
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}

func redisConveyAll(red *redis.Pool, pipeKeyPrefix string, consumers map[string]consumer.Interface) {
	var (
		pattern      = fmt.Sprintf(`%s.????????-????-????-????-????????????`, pipeKeyPrefix)
		maxTries     = 20
		retryPeriod  = 10 * time.Second
		timer        *time.Timer
		success      bool
		err          error
		scanResultsI interface{}
		scanResults  []interface{}
		cursorI      interface{}
		cursor       = 0
		pipeKeysI    interface{}
		pipeKeys     []interface{}
		pipeKeyI     interface{}
	)
	for i := 0; i < maxTries; i++ {
		for cursor != 0 || !success {
			success = false
			conn := red.Get()
			scanResultsI, err = conn.Do("SCAN", cursor, "MATCH", pattern)
			conn.Close()
			if err != nil {
				fmt.Printf("SCAN.%s; Try: %d\n", err, i)
				success = false
				break
			}
			scanResults = scanResultsI.([]interface{})
			cursorI = scanResults[0]
			cursor, err = strconv.Atoi(string(cursorI.([]byte)))
			if err != nil {
				fmt.Printf("strconv.%s; Try: %d\n", err, i)
				success = false
				break
			}
			pipeKeysI = scanResults[1]
			pipeKeys = pipeKeysI.([]interface{})
			for _, pipeKeyI = range pipeKeys {
				go redisConvey(red, string(pipeKeyI.([]byte)), consumers)
			}
			success = true
		}
		if !success {
			timer = time.NewTimer(retryPeriod)
			<-timer.C
			continue
		}
		break
	}
	if !success {
		panic(fmt.Errorf("redis KEYS kept failing after %d retries", maxTries))
	}
}
