package engine

import (
	"math"
	"sync"
	"time"

	"github.com/bulklog/bulklog/collection"
	"github.com/bulklog/bulklog/consumer"
	"github.com/bulklog/bulklog/log"
)

// convey documents to consumers through pipes!
func convey(documents []collection.Document, consumers map[string]consumer.Interface, retryPeriod, retentionPeriod time.Duration) {
	var (
		startedAt           = time.Now().UTC()
		dieAt               = startedAt.Add(retentionPeriod)
		dieAtUnixNano       = dieAt.UnixNano()
		currentTimeUnixNano int64
		nextTryAtUnixNano   int64
		i                   int
		failed              map[string]consumer.Interface
		err                 error
		timer               *time.Timer
		latestTryAt         time.Time
		waitFor             time.Duration
		cons                consumer.Interface
		consumerName        string
		wg                  sync.WaitGroup
	)
	for {
		latestTryAt = time.Now().UTC()
		wg = sync.WaitGroup{}
		for consumerName, cons = range consumers {
			wg.Add(1)
			go func(consumerName string, cons consumer.Interface) {
				err = cons.Digest(documents)
				if err != nil {
					if failed == nil {
						failed = make(map[string]consumer.Interface)
					}
					failed[consumerName] = cons
					log.Err().Printf("Digest.%s)\n", err)
				}
				wg.Done()
			}(consumerName, cons)
		}
		wg.Wait()
		if len(failed) == 0 || time.Since(startedAt) > retentionPeriod {
			return
		}
		consumers = failed
		waitFor = retryPeriod*time.Duration(math.Pow(2, float64(i))) - time.Since(latestTryAt)
		currentTimeUnixNano = time.Now().UTC().UnixNano()
		nextTryAtUnixNano = currentTimeUnixNano + int64(waitFor)
		if nextTryAtUnixNano > dieAtUnixNano || currentTimeUnixNano > dieAtUnixNano {
			return
		}
		i++
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}
