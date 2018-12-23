package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/khezen/espipe/consumer"

	"github.com/go-redis/redis"
)

func redisConvey(red redis.Client, pipeKey string) {
	startedAt, retryPeriod, retentionPeriod, err := getPipe(red, pipeKey)
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
	red redis.Client, pipeKey string,
	startedAt time.Time,
	retryPeriod, retentionPeriod time.Duration) {
	documents, err := getDocuments(red, pipeKey)
	if err != nil {
		fmt.Println(err)
		return
	}
	var (
		consumers   []consumer.Interface
		iter        int
		latestTryAt time.Time
		waitFor     time.Duration
		timer       *time.Timer
		done        bool
	)
	for {
		latestTryAt = time.Now().UTC()
		func() {
			tx := red.TxPipeline()
			defer func() {
				if err != nil {
					tx.Discard()
				} else {
					tx.Exec()
				}
			}()
			consumers, err = getConsumers(tx, pipeKey)
			if err != nil {
				fmt.Println(err)
				return
			}
			for i := 0; i < len(consumers); i++ {
				err = consumers[i].Digest(documents)
				if err != nil {
					fmt.Println(err)
					err = nil
				} else {
					consumers[i] = consumers[len(consumers)-1]
					consumers[len(consumers)-1] = nil
					consumers = consumers[:len(consumers)]
					i--
					err = setConsumers(tx, pipeKey, consumers)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
			if len(consumers) == 0 || time.Since(startedAt) > retentionPeriod {
				err = deletePipe(tx, pipeKey)
				if err != nil {
					fmt.Println(err)
				}
				done = true
				return
			}
		}()
		if done {
			return
		}
		iter, err = getIter(red, pipeKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		waitFor = retryPeriod*time.Duration(math.Pow(2, float64(iter))) - time.Since(latestTryAt)
		iter++
		err = setIter(red, pipeKey, iter)
		if err != nil {
			fmt.Println(err)
			return
		}
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}

func redisConveyAll(red redis.Client, pipeKeyPrefix string) {
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
