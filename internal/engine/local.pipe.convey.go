package engine

import (
	"math"
	"sync"
	"time"

	"github.com/khezen/bulklog/internal/collection"
	"github.com/khezen/bulklog/internal/log"
	"github.com/khezen/bulklog/internal/output"
)

// convey documents to outputs through pipes!
func convey(documents []collection.Document, outputs map[string]output.Interface, retryPeriod, retentionPeriod time.Duration) {
	var (
		startedAt           = time.Now().UTC()
		dieAt               = startedAt.Add(retentionPeriod)
		dieAtUnixNano       = dieAt.UnixNano()
		currentTimeUnixNano int64
		nextTryAtUnixNano   int64
		i                   int
		failed              map[string]output.Interface
		err                 error
		timer               *time.Timer
		latestTryAt         time.Time
		waitFor             time.Duration
		cons                output.Interface
		outputName          string
		wg                  sync.WaitGroup
	)
	for {
		latestTryAt = time.Now().UTC()
		wg = sync.WaitGroup{}
		for outputName, cons = range outputs {
			wg.Add(1)
			go func(outputName string, cons output.Interface) {
				err = cons.Digest(documents)
				if err != nil {
					if failed == nil {
						failed = make(map[string]output.Interface)
					}
					failed[outputName] = cons
					log.Err().Printf("Digest.%s)\n", err)
				}
				wg.Done()
			}(outputName, cons)
		}
		wg.Wait()
		if len(failed) == 0 || time.Since(startedAt) > retentionPeriod {
			return
		}
		outputs = failed
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
