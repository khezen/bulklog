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
		startedAt   = time.Now().UTC()
		i           int
		failed      []consumer.Interface
		err         error
		timer       *time.Timer
		latestTryAt time.Time
		waitFor     time.Duration
		cons        consumer.Interface
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
		i++
		if waitFor <= 0 {
			continue
		}
		timer = time.NewTimer(waitFor)
		<-timer.C
	}
}
