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

	throttleCounter = metrics.NewRegisteredCounter("sil/downloader/throttle", nil)

	// snapPeerSkipMeter tracks snap peers skipped by the state syncer because
	// they negotiated a version below the one the syncer requires.
	snapPeerSkipMeter = metrics.NewRegisteredMeter("sil/downloader/snap/peerskip", nil)
)
