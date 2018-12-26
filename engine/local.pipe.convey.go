package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/consumer"
)

// convey documents to consumers through pipes!
func convey(documents []collection.Document, consumers []consumer.Interface, retryPeriod, retentionPeriod time.Duration) {
	var (
		startedAt           = time.Now().UTC()
		dieAt               = startedAt.Add(retentionPeriod)
		dieAtUnixNano       = dieAt.UnixNano()
		currentTimeUnixNano int64
		nextTryAtUnixNano   int64
		i                   int
		failed              []consumer.Interface
		err                 error
		timer               *time.Timer
		latestTryAt         time.Time
		waitFor             time.Duration
		cons                consumer.Interface
	)
	for {
		latestTryAt = time.Now().UTC()
		for _, cons = range consumers {
			err = cons.Digest(documents)
			if err != nil {
				if failed == nil {
					failed = make([]consumer.Interface, 0, len(consumers))
				}
				failed = append(failed, cons)
				fmt.Println(err)
			}
		}
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
