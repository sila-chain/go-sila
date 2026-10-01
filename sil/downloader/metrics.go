// Copyright 2015 The go-sila Authors
// This file is part of the go-sila library.
//
// The go-sila library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-sila library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-sila library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/sila-chain/go-sila/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("sil/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("sil/downloader/headers/req", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("sil/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("sil/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("sil/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("sil/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("sil/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("sil/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("sil/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("sil/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("sil/downloader/receipts/timeout", nil)

	balInMeter      = metrics.NewRegisteredMeter("sil/downloader/bals/in", nil)
	balReqTimer     = metrics.NewRegisteredTimer("sil/downloader/bals/req", nil)
	balDropMeter    = metrics.NewRegisteredMeter("sil/downloader/bals/drop", nil)
	balTimeoutMeter = metrics.NewRegisteredMeter("sil/downloader/bals/timeout", nil)

	bodyFetchMetrics    = newFetchMetrics("bodies")
	receiptFetchMetrics = newFetchMetrics("receipts")
	balFetchMetrics     = newFetchMetrics("bals")

	// Chain download progress, reported alongside the progress log
	chainProgressGauge = metrics.NewRegisteredGaugeFloat64("sil/downloader/chain/progress", nil)

	// rttTargetGauge is the round trip time (in milliseconds) requests are
	// currently sized for, derived from the median of the peer estimates.
	rttTargetGauge = metrics.NewRegisteredGauge("sil/downloader/rtt/target", nil)

	importWaitTimer           = metrics.NewRegisteredTimer("sil/downloader/import/wait", nil)
	importInsertBlocksTimer   = metrics.NewRegisteredTimer("sil/downloader/import/blocks", nil)
	importInsertReceiptsTimer = metrics.NewRegisteredTimer("sil/downloader/import/receipts", nil)
	importBatchHistogram      = metrics.NewRegisteredHistogram("sil/downloader/import/batch", nil, metrics.NewExpDecaySample(1028, 0.015))

	// Result cache metrics, reported every time a batch is drained.
	queueThrottleGauge = metrics.NewRegisteredGauge("sil/downloader/queue/throttle/threshold", nil)
	queueItemSizeGauge = metrics.NewRegisteredGauge("sil/downloader/queue/itemsize", nil)

	// snapPeerSkipMeter tracks snap peers skipped by the state syncer because
	// they negotiated a version below the one the syncer requires.
	snapPeerSkipMeter = metrics.NewRegisteredMeter("sil/downloader/snap/peerskip", nil)
)

// fetchMetrics groups the collectors the concurrent fetcher reports into for a
// single data type (bodies, receipts, access lists).
type fetchMetrics struct {
	idlePeers    *metrics.Gauge    // Peers left without a request after an assignment round
	busyPeers    *metrics.Gauge    // Peers with a request in flight
	stalePeers   *metrics.Gauge    // Peers with a timed out but not yet answered request
	rangedPeers  *metrics.Gauge    // Peers whose announced block range excludes the next block to hand out
	capacity     *metrics.Gauge    // Estimated aggregate items per second across all peers
	starved      *metrics.Meter    // Assignment rounds cut short because nothing was pending
	throttled    *metrics.Meter    // Assignment rounds cut short by result cache throttling
	headExpiries *metrics.Meter    // Requests expired early for holding the result cache head
	items        metrics.Histogram // Items contained in each response
	bytes        *metrics.Meter    // Payload bytes contained in each response
}

// newFetchMetrics registers the scheduling collectors for a data type under
// sil/downloader/<kind>/...
func newFetchMetrics(kind string) *fetchMetrics {
	prefix := "sil/downloader/" + kind
	return &fetchMetrics{
		idlePeers:    metrics.NewRegisteredGauge(prefix+"/peers/idle", nil),
		busyPeers:    metrics.NewRegisteredGauge(prefix+"/peers/busy", nil),
		stalePeers:   metrics.NewRegisteredGauge(prefix+"/peers/stale", nil),
		rangedPeers:  metrics.NewRegisteredGauge(prefix+"/peers/outofrange", nil),
		capacity:     metrics.NewRegisteredGauge(prefix+"/capacity", nil),
		starved:      metrics.NewRegisteredMeter(prefix+"/starved", nil),
		throttled:    metrics.NewRegisteredMeter(prefix+"/throttled", nil),
		headExpiries: metrics.NewRegisteredMeter(prefix+"/headexpire", nil),
		items:        metrics.NewRegisteredHistogram(prefix+"/items", nil, metrics.NewExpDecaySample(1028, 0.015)),
		bytes:        metrics.NewRegisteredMeter(prefix+"/bytes", nil),
	}
}
