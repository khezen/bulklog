package engine

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/bulklog/bulklog/log"
	"github.com/bulklog/bulklog/output"
	"github.com/gomodule/redigo/redis"
)

func redisConvey(red *redis.Pool, pipeKey string, outputs map[string]output.Interface) {
	startedAt, retryPeriod, retentionPeriod, err := getRedisPipe(red, pipeKey)
	if err == errRedisPipeNotFound {
		err = deleteRedisPipe(red, pipeKey)
		if err != nil {
			log.Err().Printf("deleteRedisPipe.%s)\n", err)
		}
		return
	}
	if err != nil {
		log.Err().Printf("getRedisPipe.%s)\n", err)
		return
	}
	presetRedisConvey(
		red, pipeKey,
		outputs,
		startedAt,
		retryPeriod, retentionPeriod,
	)
}

func presetRedisConvey(
	red *redis.Pool, pipeKey string,
	outputs map[string]output.Interface,
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
			log.Err().Printf("deleteRedisPipe.%s)\n", err)
		} else {
			return
		}
	}
	documents, err := getRedisPipeDocuments(red, pipeKey)
	if err != nil {
		log.Err().Printf("getRedisPipeDocuments.%s)\n", err)
		return
	}
	if len(documents) == 0 {
		err = deleteRedisPipe(red, pipeKey)
		if err != nil {
			log.Err().Printf("deleteRedisPipe.%s)\n", err)
		}
		return
	}
	var (
		remainingoutputs  map[string]output.Interface
		nextTryAtUnixNano int64
		iteration         int
		latestTryAt       time.Time
		waitFor           time.Duration
		timer             *time.Timer
		wg                sync.WaitGroup
	)
	for {
		latestTryAt = time.Now().UTC()
		remainingoutputs, err = getRedisPipeoutputs(red, pipeKey, outputs)
		if err != nil {
			log.Err().Printf("getRedisPipeoutputs.%s)\n", err)
			return
		}
		if len(remainingoutputs) == 0 {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				log.Err().Printf("deleteRedisPipe.%s)\n", err)
			}
			return
		}
		wg = sync.WaitGroup{}
		for outputName, cons := range remainingoutputs {
			wg.Add(1)
			go func(outputName string, cons output.Interface) {
				err = cons.Digest(documents)
				if err != nil {
					log.Err().Printf("Digest.%s)\n", err)
					err = nil
				} else {
					err = deleteRedisPipeoutput(red, pipeKey, outputName)
					if err != nil {
						log.Err().Printf("deleteRedisPipeoutput.%s)\n", err)
					}
				}
				wg.Done()
			}(outputName, cons)
		}
		wg.Wait()
		currentTimeUnixNano = time.Now().UTC().UnixNano()
		if len(remainingoutputs) == 0 || currentTimeUnixNano > dieAtUnixNano {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				log.Err().Printf("deleteRedisPipe.%s)\n", err)
			} else {
				return
			}
		}
		iteration, err = getRedisPipeIteration(red, pipeKey)
		if err != nil {
			log.Err().Printf("getRedisPipeIteration.%s)\n", err)
			waitFor = retryPeriod - time.Since(latestTryAt)
			continue
		}
		waitFor = retryPeriod*time.Duration(math.Pow(2, float64(iteration))) - time.Since(latestTryAt)
		nextTryAtUnixNano = currentTimeUnixNano + int64(waitFor)
		if nextTryAtUnixNano > dieAtUnixNano {
			err = deleteRedisPipe(red, pipeKey)
			if err != nil {
				log.Err().Printf("deleteRedisPipe.%s)\n", err)
			} else {
				return
			}
		}
		err = incrRedisPipeIteration(red, pipeKey)
		if err != nil {
			log.Err().Printf("incrRedisPipeIteration.%s)\n", err)
			return
		}
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}

func redisConveyAll(red *redis.Pool, pipeKeyPrefix string, outputs map[string]output.Interface) {
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
				log.Err().Printf("SCAN.%s; Try: %d\n", err, i)
				success = false
				break
			}
			scanResults = scanResultsI.([]interface{})
			cursorI = scanResults[0]
			cursor, err = strconv.Atoi(string(cursorI.([]byte)))
			if err != nil {
				log.Err().Printf("strconv.%s; Try: %d\n", err, i)
				success = false
				break
			}
			pipeKeysI = scanResults[1]
			pipeKeys = pipeKeysI.([]interface{})
			for _, pipeKeyI = range pipeKeys {
				go redisConvey(red, string(pipeKeyI.([]byte)), outputs)
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
