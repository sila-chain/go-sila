// Copyright 2023 The go-sila Authors
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

package vm

import (
	"errors"

	"github.com/sila-chain/go-sila/params"
)

// LookupInstructionSet returns the instruction set for the fork configured by
// the rules.
func LookupInstructionSet(rules params.Rules) (JumpTable, error) {
	switch {
	case rules.IsUBT:
		return newSilaCancunInstructionSet(), errors.New("verkle-fork not defined yet")
	case rules.IsAmsterdam:
		return newAmsterdamInstructionSet(), nil
	case rules.IsSilaOsaka:
		return newSilaOsakaInstructionSet(), nil
	case rules.IsSilaPrague:
		return newSilaPragueInstructionSet(), nil
	case rules.IsSilaCancun:
		return newSilaCancunInstructionSet(), nil
	case rules.IsSilaShanghai:
		return newSilaShanghaiInstructionSet(), nil
	case rules.IsMerge:
		return newMergeInstructionSet(), nil
	case rules.IsSilaLondon:
		return newSilaLondonInstructionSet(), nil
	case rules.IsSilaBerlin:
		return newSilaBerlinInstructionSet(), nil
	case rules.IsSilaIstanbul:
		return newSilaIstanbulInstructionSet(), nil
	case rules.IsSilaConstantinople:
		return newSilaConstantinopleInstructionSet(), nil
	case rules.IsSilaByzantium:
		return newSilaByzantiumInstructionSet(), nil
	case rules.IsSIP158:
		return newSpuriousDragonInstructionSet(), nil
	case rules.IsSIP150:
		return newTangerineWhistleInstructionSet(), nil
	case rules.IsSilaHomestead:
		return newSilaHomesteadInstructionSet(), nil
	}
	return newFrontierInstructionSet(), nil
}

// Stack returns the minimum and maximum stack requirements.
func (op *operation) Stack() (int, int) {
	return op.minStack, op.maxStack
}

// HasCost returns true if the opcode has a cost. Opcodes which do _not_ have
// a cost assigned are one of two things:
// - undefined, a.k.a invalid opcodes,
// - the STOP opcode.
// This method can thus be used to check if an opcode is "Invalid (or STOP)".
func (op *operation) HasCost() bool {
	// Ideally, we'd check this:
	//	return op.execute == opUndefined
	// However, go-lang does now allow that. So we'll just check some other
	// 'indicators' that this is an invalid op. Alas, STOP is impossible to
	// filter out
	return op.dynamicGas != nil || op.constantGas != 0
}
