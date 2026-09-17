// Copyright 2025 The go-sila Authors
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
// along with the go-sila library. If not, see <http://www.gnu.org/licenses/

// Contains the metrics collected by the txfetcher.

package fetcher

import "github.com/sila-chain/go-sila/metrics"

var (
	txAnnounceInMeter          = metrics.NewRegisteredMeter("sil/fetcher/transaction/announces/in", nil)
	txAnnounceKnownMeter       = metrics.NewRegisteredMeter("sil/fetcher/transaction/announces/known", nil)
	txAnnounceUnderpricedMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/announces/underpriced", nil)
	txAnnounceOnchainMeter     = metrics.NewRegisteredMeter("sil/fetcher/transaction/announces/onchain", nil)
	txAnnounceDOSMeter         = metrics.NewRegisteredMeter("sil/fetcher/transaction/announces/dos", nil)

	txBroadcastInMeter          = metrics.NewRegisteredMeter("sil/fetcher/transaction/broadcasts/in", nil)
	txBroadcastKnownMeter       = metrics.NewRegisteredMeter("sil/fetcher/transaction/broadcasts/known", nil)
	txBroadcastUnderpricedMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/broadcasts/underpriced", nil)
	txBroadcastOtherRejectMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/broadcasts/otherreject", nil)

	txRequestOutMeter     = metrics.NewRegisteredMeter("sil/fetcher/transaction/request/out", nil)
	txRequestFailMeter    = metrics.NewRegisteredMeter("sil/fetcher/transaction/request/fail", nil)
	txRequestDoneMeter    = metrics.NewRegisteredMeter("sil/fetcher/transaction/request/done", nil)
	txRequestTimeoutMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/request/timeout", nil)

	txReplyInMeter          = metrics.NewRegisteredMeter("sil/fetcher/transaction/replies/in", nil)
	txReplyKnownMeter       = metrics.NewRegisteredMeter("sil/fetcher/transaction/replies/known", nil)
	txReplyUnderpricedMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/replies/underpriced", nil)
	txReplyOtherRejectMeter = metrics.NewRegisteredMeter("sil/fetcher/transaction/replies/otherreject", nil)

	txFetcherWaitingPeers   = metrics.NewRegisteredGauge("sil/fetcher/transaction/waiting/peers", nil)
	txFetcherWaitingHashes  = metrics.NewRegisteredGauge("sil/fetcher/transaction/waiting/hashes", nil)
	txFetcherQueueingPeers  = metrics.NewRegisteredGauge("sil/fetcher/transaction/queueing/peers", nil)
	txFetcherQueueingHashes = metrics.NewRegisteredGauge("sil/fetcher/transaction/queueing/hashes", nil)
	txFetcherFetchingPeers  = metrics.NewRegisteredGauge("sil/fetcher/transaction/fetching/peers", nil)
	txFetcherFetchingHashes = metrics.NewRegisteredGauge("sil/fetcher/transaction/fetching/hashes", nil)

	txFetcherSlowPeers = metrics.NewRegisteredGauge("sil/fetcher/transaction/slow/peers", nil)
	// Note: this metric does not mean that the fetching of a transaction
	// was blocked by a specific peer during this period, since we request
	// another peer to fetch the same transaction hash.
	// The purpose of this metric is to measure how long it takes for a slow peer
	// to become "unfrozen", either by eventually replying to the request
	// or by being dropped, measuring from the moment the request was sent.
	txFetcherSlowWait = metrics.NewRegisteredHistogram("eth/fetcher/transaction/slow/wait", nil, metrics.NewExpDecaySample(1028, 0.015))

	blobAnnounceInMeter  = metrics.NewRegisteredMeter("eth/fetcher/blob/announces/in", nil)
	blobAnnounceDOSMeter = metrics.NewRegisteredMeter("eth/fetcher/blob/announces/dos", nil)
	// This metric tracks partial→full conversions due to availability timeout
	blobAnnounceTimeoutMeter = metrics.NewRegisteredMeter("eth/fetcher/blob/announces/timeout", nil)

	blobRequestOutMeter     = metrics.NewRegisteredMeter("eth/fetcher/blob/request/out", nil)
	blobRequestFailMeter    = metrics.NewRegisteredMeter("eth/fetcher/blob/request/fail", nil)
	blobRequestDoneMeter    = metrics.NewRegisteredMeter("eth/fetcher/blob/request/done", nil)
	blobRequestTimeoutMeter = metrics.NewRegisteredMeter("eth/fetcher/blob/request/timeout", nil)

	blobReplyInMeter = metrics.NewRegisteredMeter("eth/fetcher/blob/replies/in", nil)

	blobFetcherWaitingPeers   = metrics.NewRegisteredGauge("eth/fetcher/blob/waiting/peers", nil)
	blobFetcherWaitingHashes  = metrics.NewRegisteredGauge("eth/fetcher/blob/waiting/hashes", nil)
	blobFetcherQueueingPeers  = metrics.NewRegisteredGauge("eth/fetcher/blob/queueing/peers", nil)
	blobFetcherQueueingHashes = metrics.NewRegisteredGauge("eth/fetcher/blob/queueing/hashes", nil)
	blobFetcherFetchingPeers  = metrics.NewRegisteredGauge("eth/fetcher/blob/fetching/peers", nil)
	blobFetcherFetchingHashes = metrics.NewRegisteredGauge("eth/fetcher/blob/fetching/hashes", nil)

	blobFetcherWaitTime  = metrics.NewRegisteredHistogram("eth/fetcher/blob/wait/time", nil, metrics.NewExpDecaySample(1028, 0.015))
	blobFetcherFetchTime = metrics.NewRegisteredHistogram("eth/fetcher/blob/fetch/time", nil, metrics.NewExpDecaySample(1028, 0.015))
)
