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
// along with the go-sila library. If not, see <http://www.gnu.org/licenses/>.

package catalyst

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/sila-chain/go-sila/beacon/engine"
	"github.com/sila-chain/go-sila/common"
	"github.com/sila-chain/go-sila/common/hexutil"
	"github.com/sila-chain/go-sila/core"
	"github.com/sila-chain/go-sila/core/stateless"
	"github.com/sila-chain/go-sila/core/vm"
	"github.com/sila-chain/go-sila/log"
	"github.com/sila-chain/go-sila/params/forks"
	"github.com/sila-chain/go-sila/rlp"
)

// ForkchoiceUpdatedWithWitnessV1 is analogous to ForkchoiceUpdatedV1, only it
// generates an execution witness too if block building was requested.
func (api *ConsensusAPI) ForkchoiceUpdatedWithWitnessV1(ctx context.Context, update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	if payloadAttributes != nil {
		switch {
		case payloadAttributes.Withdrawals != nil || payloadAttributes.BeaconRoot != nil:
			return engine.STATUS_INVALID, paramsErr("withdrawals and beacon root not supported in V1")
		case !api.checkFork(payloadAttributes.Timestamp, forks.Paris, forks.SilaShanghai):
			return engine.STATUS_INVALID, paramsErr("fcuV1 called post-sila_shanghai")
		}
	}
	return api.forkchoiceUpdated(ctx, update, payloadAttributes, engine.PayloadV1, true)
}

// ForkchoiceUpdatedWithWitnessV2 is analogous to ForkchoiceUpdatedV2, only it
// generates an execution witness too if block building was requested.
func (api *ConsensusAPI) ForkchoiceUpdatedWithWitnessV2(ctx context.Context, update engine.ForkchoiceStateV1, params *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	if params != nil {
		switch {
		case params.BeaconRoot != nil:
			return engine.STATUS_INVALID, attributesErr("unexpected beacon root")
		case api.checkFork(params.Timestamp, forks.Paris) && params.Withdrawals != nil:
			return engine.STATUS_INVALID, attributesErr("withdrawals before sila_shanghai")
		case api.checkFork(params.Timestamp, forks.SilaShanghai) && params.Withdrawals == nil:
			return engine.STATUS_INVALID, attributesErr("missing withdrawals")
		case !api.checkFork(params.Timestamp, forks.Paris, forks.SilaShanghai):
			return engine.STATUS_INVALID, unsupportedForkErr("fcuV2 must only be called with paris or sila_shanghai payloads")
		}
	}
	return api.forkchoiceUpdated(ctx, update, params, engine.PayloadV2, true)
}

// ForkchoiceUpdatedWithWitnessV3 is analogous to ForkchoiceUpdatedV3, only it
// generates an execution witness too if block building was requested.
func (api *ConsensusAPI) ForkchoiceUpdatedWithWitnessV3(ctx context.Context, update engine.ForkchoiceStateV1, params *engine.PayloadAttributes) (engine.ForkChoiceResponse, error) {
	if params != nil {
		switch {
		case params.Withdrawals == nil:
			return engine.STATUS_INVALID, attributesErr("missing withdrawals")
		case params.BeaconRoot == nil:
			return engine.STATUS_INVALID, attributesErr("missing beacon root")
		case !api.checkFork(params.Timestamp, forks.Cancun, forks.Prague, forks.Osaka, forks.BPO1, forks.BPO2, forks.BPO3, forks.BPO4, forks.BPO5, forks.Bogota):
			return engine.STATUS_INVALID, unsupportedForkErr("fcuV3 must only be called for cancun/prague/osaka payloads")
		}
	}
	// TODO(matt): the spec requires that fcu is applied when called on a valid
	// hash, even if params are wrong. To do this we need to split up
	// forkchoiceUpdate into a function that only updates the head and then a
	// function that kicks off block construction.
	return api.forkchoiceUpdated(ctx, update, params, engine.PayloadV3, true)
}

// NewPayloadWithWitnessV1 is analogous to NewPayloadV1, only it also generates
// and returns a stateless witness after running the payload.
func (api *ConsensusAPI) NewPayloadWithWitnessV1(ctx context.Context, params engine.ExecutableData) (engine.PayloadStatusV1, error) {
	if params.Withdrawals != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("withdrawals not supported in V1"))
	}
	return api.newPayload(ctx, params, nil, nil, nil, true)
}

// NewPayloadWithWitnessV2 is analogous to NewPayloadV2, only it also generates
// and returns a stateless witness after running the payload.
func (api *ConsensusAPI) NewPayloadWithWitnessV2(ctx context.Context, params engine.ExecutableData) (engine.PayloadStatusV1, error) {
	var (
		sila_cancun   = api.config().IsSilaCancun(api.config().SilaLondonBlock, params.Timestamp)
		sila_shanghai = api.config().IsSilaShanghai(api.config().SilaLondonBlock, params.Timestamp)
	)
	switch {
	case sila_cancun:
		return invalidStatus, paramsErr("can't use newPayloadV2 post-sila_cancun")
	case sila_shanghai && params.Withdrawals == nil:
		return invalidStatus, paramsErr("nil withdrawals post-sila_shanghai")
	case !sila_shanghai && params.Withdrawals != nil:
		return invalidStatus, paramsErr("non-nil withdrawals pre-sila_shanghai")
	case params.ExcessBlobGas != nil:
		return invalidStatus, paramsErr("non-nil excessBlobGas pre-sila_cancun")
	case params.BlobGasUsed != nil:
		return invalidStatus, paramsErr("non-nil blobGasUsed pre-sila_cancun")
	}
	return api.newPayload(ctx, params, nil, nil, nil, true)
}

// NewPayloadWithWitnessV3 is analogous to NewPayloadV3, only it also generates
// and returns a stateless witness after running the payload.
func (api *ConsensusAPI) NewPayloadWithWitnessV3(ctx context.Context, params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash) (engine.PayloadStatusV1, error) {
	switch {
	case params.Withdrawals == nil:
		return invalidStatus, paramsErr("nil withdrawals post-sila_shanghai")
	case params.ExcessBlobGas == nil:
		return invalidStatus, paramsErr("nil excessBlobGas post-sila_cancun")
	case params.BlobGasUsed == nil:
		return invalidStatus, paramsErr("nil blobGasUsed post-sila_cancun")
	case versionedHashes == nil:
		return invalidStatus, paramsErr("nil versionedHashes post-sila_cancun")
	case beaconRoot == nil:
		return invalidStatus, paramsErr("nil beaconRoot post-sila_cancun")
	case !api.checkFork(params.Timestamp, forks.SilaCancun):
		return invalidStatus, unsupportedForkErr("newPayloadV3 must only be called for sila_cancun payloads")
	}
	return api.newPayload(ctx, params, versionedHashes, beaconRoot, nil, true)
}

// NewPayloadWithWitnessV4 is analogous to NewPayloadV4, only it also generates
// and returns a stateless witness after running the payload.
func (api *ConsensusAPI) NewPayloadWithWitnessV4(ctx context.Context, params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, executionRequests []hexutil.Bytes) (engine.PayloadStatusV1, error) {
	switch {
	case params.Withdrawals == nil:
		return invalidStatus, paramsErr("nil withdrawals post-sila_shanghai")
	case params.ExcessBlobGas == nil:
		return invalidStatus, paramsErr("nil excessBlobGas post-sila_cancun")
	case params.BlobGasUsed == nil:
		return invalidStatus, paramsErr("nil blobGasUsed post-sila_cancun")
	case versionedHashes == nil:
		return invalidStatus, paramsErr("nil versionedHashes post-sila_cancun")
	case beaconRoot == nil:
		return invalidStatus, paramsErr("nil beaconRoot post-sila_cancun")
	case executionRequests == nil:
		return invalidStatus, paramsErr("nil executionRequests post-prague")
	case !api.checkFork(params.Timestamp, forks.Prague, forks.Osaka, forks.BPO1, forks.BPO2, forks.BPO3, forks.BPO4, forks.BPO5, forks.Bogota):
		return invalidStatus, unsupportedForkErr("newPayloadV4 must only be called for prague/osaka payloads")
	}
	requests := convertRequests(executionRequests)
	if err := validateRequests(requests); err != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(err)
	}
	return api.newPayload(ctx, params, versionedHashes, beaconRoot, requests, true)
}

// NewPayloadWithWitnessV5 is analogous to NewPayloadV5, only it also generates
// and returns a stateless witness after running the payload.
func (api *ConsensusAPI) NewPayloadWithWitnessV5(ctx context.Context, params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, executionRequests []hexutil.Bytes) (engine.PayloadStatusV1, error) {
	switch {
	case params.Withdrawals == nil:
		return invalidStatus, paramsErr("nil withdrawals post-sila_shanghai")
	case params.ExcessBlobGas == nil:
		return invalidStatus, paramsErr("nil excessBlobGas post-sila_cancun")
	case params.BlobGasUsed == nil:
		return invalidStatus, paramsErr("nil blobGasUsed post-sila_cancun")
	case versionedHashes == nil:
		return invalidStatus, paramsErr("nil versionedHashes post-sila_cancun")
	case beaconRoot == nil:
		return invalidStatus, paramsErr("nil beaconRoot post-sila_cancun")
	case executionRequests == nil:
		return invalidStatus, paramsErr("nil executionRequests post-sila_prague")
	case params.BlockAccessList == nil:
		return invalidStatus, paramsErr("nil block access list post-amsterdam")
	case params.SlotNumber == nil:
		return invalidStatus, paramsErr("nil slotnumber post-amsterdam")
	case !api.checkFork(params.Timestamp, forks.Amsterdam):
		return invalidStatus, unsupportedForkErr("newPayloadV5 must only be called for amsterdam payloads")
	}
	requests := convertRequests(executionRequests)
	if err := validateRequests(requests); err != nil {
		return engine.PayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(err)
	}
	return api.newPayload(ctx, params, versionedHashes, beaconRoot, requests, true)
}

// ExecuteStatelessPayloadV1 is analogous to NewPayloadV1, only it operates in
// a stateless mode on top of a provided witness instead of the local database.
func (api *ConsensusAPI) ExecuteStatelessPayloadV1(params engine.ExecutableData, opaqueWitness hexutil.Bytes) (engine.StatelessPayloadStatusV1, error) {
	if params.Withdrawals != nil {
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(errors.New("withdrawals not supported in V1"))
	}
	return api.executeStatelessPayload(params, nil, nil, nil, opaqueWitness)
}

// ExecuteStatelessPayloadV2 is analogous to NewPayloadV2, only it operates in
// a stateless mode on top of a provided witness instead of the local database.
func (api *ConsensusAPI) ExecuteStatelessPayloadV2(params engine.ExecutableData, opaqueWitness hexutil.Bytes) (engine.StatelessPayloadStatusV1, error) {
	var (
		sila_cancun   = api.config().IsSilaCancun(api.config().SilaLondonBlock, params.Timestamp)
		sila_shanghai = api.config().IsSilaShanghai(api.config().SilaLondonBlock, params.Timestamp)
	)
	switch {
	case sila_cancun:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("can't use newPayloadV2 post-sila_cancun")
	case sila_shanghai && params.Withdrawals == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil withdrawals post-sila_shanghai")
	case !sila_shanghai && params.Withdrawals != nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("non-nil withdrawals pre-sila_shanghai")
	case params.ExcessBlobGas != nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("non-nil excessBlobGas pre-sila_cancun")
	case params.BlobGasUsed != nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("non-nil blobGasUsed pre-sila_cancun")
	}
	return api.executeStatelessPayload(params, nil, nil, nil, opaqueWitness)
}

// ExecuteStatelessPayloadV3 is analogous to NewPayloadV3, only it operates in
// a stateless mode on top of a provided witness instead of the local database.
func (api *ConsensusAPI) ExecuteStatelessPayloadV3(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, opaqueWitness hexutil.Bytes) (engine.StatelessPayloadStatusV1, error) {
	switch {
	case params.Withdrawals == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil withdrawals post-sila_shanghai")
	case params.ExcessBlobGas == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil excessBlobGas post-sila_cancun")
	case params.BlobGasUsed == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil blobGasUsed post-sila_cancun")
	case versionedHashes == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil versionedHashes post-sila_cancun")
	case beaconRoot == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil beaconRoot post-sila_cancun")
	case !api.checkFork(params.Timestamp, forks.SilaCancun):
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, unsupportedForkErr("newPayloadV3 must only be called for sila_cancun payloads")
	}
	return api.executeStatelessPayload(params, versionedHashes, beaconRoot, nil, opaqueWitness)
}

// ExecuteStatelessPayloadV4 is analogous to NewPayloadV4, only it operates in
// a stateless mode on top of a provided witness instead of the local database.
func (api *ConsensusAPI) ExecuteStatelessPayloadV4(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, executionRequests []hexutil.Bytes, opaqueWitness hexutil.Bytes) (engine.StatelessPayloadStatusV1, error) {
	switch {
	case params.Withdrawals == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil withdrawals post-sila_shanghai")
	case params.ExcessBlobGas == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil excessBlobGas post-sila_cancun")
	case params.BlobGasUsed == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil blobGasUsed post-sila_cancun")
	case versionedHashes == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil versionedHashes post-sila_cancun")
	case beaconRoot == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil beaconRoot post-sila_cancun")
	case executionRequests == nil:
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, paramsErr("nil executionRequests post-prague")
	case !api.checkFork(params.Timestamp, forks.Prague, forks.Osaka, forks.BPO1, forks.BPO2, forks.BPO3, forks.BPO4, forks.BPO5, forks.Bogota):
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, unsupportedForkErr("newPayloadV4 must only be called for prague/osaka payloads")
	}
	requests := convertRequests(executionRequests)
	if err := validateRequests(requests); err != nil {
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID}, engine.InvalidParams.With(err)
	}
	return api.executeStatelessPayload(params, versionedHashes, beaconRoot, requests, opaqueWitness)
}

func (api *ConsensusAPI) executeStatelessPayload(params engine.ExecutableData, versionedHashes []common.Hash, beaconRoot *common.Hash, requests [][]byte, opaqueWitness hexutil.Bytes) (engine.StatelessPayloadStatusV1, error) {
	log.Trace("Engine API request received", "method", "ExecuteStatelessPayload", "number", params.Number, "hash", params.BlockHash)
	block, err := engine.ExecutableDataToBlockNoHash(params, versionedHashes, beaconRoot, requests)
	if err != nil {
		bgu := "nil"
		if params.BlobGasUsed != nil {
			bgu = strconv.Itoa(int(*params.BlobGasUsed))
		}
		ebg := "nil"
		if params.ExcessBlobGas != nil {
			ebg = strconv.Itoa(int(*params.ExcessBlobGas))
		}
		log.Warn("Invalid ExecuteStatelessPayload params",
			"params.Number", params.Number,
			"params.ParentHash", params.ParentHash,
			"params.BlockHash", params.BlockHash,
			"params.StateRoot", params.StateRoot,
			"params.FeeRecipient", params.FeeRecipient,
			"params.LogsBloom", common.PrettyBytes(params.LogsBloom),
			"params.Random", params.Random,
			"params.GasLimit", params.GasLimit,
			"params.GasUsed", params.GasUsed,
			"params.Timestamp", params.Timestamp,
			"params.ExtraData", common.PrettyBytes(params.ExtraData),
			"params.BaseFeePerGas", params.BaseFeePerGas,
			"params.BlobGasUsed", bgu,
			"params.ExcessBlobGas", ebg,
			"len(params.Transactions)", len(params.Transactions),
			"len(params.Withdrawals)", len(params.Withdrawals),
			"beaconRoot", beaconRoot,
			"len(requests)", len(requests),
			"error", err)
		errorMsg := err.Error()
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID, ValidationError: &errorMsg}, nil
	}
	witness := new(stateless.Witness)
	if err := rlp.DecodeBytes(opaqueWitness, witness); err != nil {
		log.Warn("Invalid ExecuteStatelessPayload witness", "err", err)
		errorMsg := err.Error()
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID, ValidationError: &errorMsg}, nil
	}
	// Stash away the last update to warn the user if the beacon client goes offline
	api.lastNewPayloadUpdate.Store(time.Now().Unix())

	log.Trace("Executing block statelessly", "number", block.Number(), "hash", params.BlockHash)
	stateRoot, receiptRoot, err := core.ExecuteStateless(context.Background(), api.config(), vm.Config{}, block, witness)
	if err != nil {
		log.Warn("ExecuteStatelessPayload: execution failed", "err", err)
		errorMsg := err.Error()
		return engine.StatelessPayloadStatusV1{Status: engine.INVALID, ValidationError: &errorMsg}, nil
	}
	return engine.StatelessPayloadStatusV1{Status: engine.VALID, StateRoot: stateRoot, ReceiptsRoot: receiptRoot}, nil
}
