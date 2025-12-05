/* SPDX-FileCopyrightText: © 2019-2022 Nadim Kobeissi <nadim@symbolic.software>
 * SPDX-License-Identifier: GPL-3.0-only */
// 00000000000000000000000000000000

package vplogic

import (
	"sync"
)

var attackerStateShared AttackerState
var attackerStateMutex sync.Mutex

func attackerStateInit(active bool) {
	attackerStateInitQuantum(active, false)
}

func attackerStateInitQuantum(active bool, quantum bool) {
	attackerStateMutex.Lock()
	attackerStateShared = AttackerState{
		Active:         active,
		Quantum:        quantum,
		CurrentPhase:   0,
		Exhausted:      false,
		Known:          []*Value{},
		PrincipalState: []*PrincipalState{},
	}
	attackerStateMutex.Unlock()
}

func attackerStatePutKnownLocked(known *Value, valPrincipalState *PrincipalState) bool {
	if valueEquivalentValueInValues(known, attackerStateShared.Known) >= 0 {
		return false
	}
	valPrincipalStateClone := constructPrincipalStateClone(valPrincipalState, false)
	attackerStateShared.Known = append(attackerStateShared.Known, known)
	attackerStateShared.PrincipalState = append(
		attackerStateShared.PrincipalState, valPrincipalStateClone,
	)
	if attackerStateShared.Quantum {
		attackerStateQuantumAbsorbLocked(known, valPrincipalState)
	}
	return true
}

func attackerStateQuantumAbsorbLocked(known *Value, valPrincipalState *PrincipalState) {
	if known.Kind != typesEnumEquation {
		return
	}
	eq := valueFlattenEquation(known.Data.(*Equation))
	if len(eq.Values) < 2 || !valueEquivalentValues(eq.Values[0], valueG, true) {
		return
	}
	for _, exponent := range eq.Values[1:] {
		attackerStatePutKnownLocked(exponent, valPrincipalState)
	}
}

func attackerStateAbsorbPhaseValues(valKnowledgeMap *KnowledgeMap, valPrincipalState *PrincipalState) error {
	attackerStateMutex.Lock()
	for i := 0; i < len(valPrincipalState.Constants); i++ {
		switch valPrincipalState.Assigned[i].Kind {
		case typesEnumConstant:
			if valPrincipalState.Assigned[i].Data.(*Constant).Qualifier != typesEnumPublic {
				continue
			}
			earliestPhase, err := minIntInSlice(valPrincipalState.Phase[i])
			if err == nil && earliestPhase > attackerStateShared.CurrentPhase {
				continue
			}
			if !valueConstantIsUsedByAtLeastOnePrincipalInKnowledgeMap(
				valKnowledgeMap, valPrincipalState.Assigned[i].Data.(*Constant),
			) {
				continue
			}
			attackerStatePutKnownLocked(valPrincipalState.Assigned[i], valPrincipalState)
		}
	}
	for i, c := range valPrincipalState.Constants {
		cc := &Value{Kind: typesEnumConstant, Data: c}
		a := valPrincipalState.Assigned[i]
		if len(valPrincipalState.Wire[i]) == 0 && !valPrincipalState.Constants[i].Leaked {
			continue
		}
		if valPrincipalState.Constants[i].Qualifier == typesEnumPublic {
			continue
		}
		earliestPhase, err := minIntInSlice(valPrincipalState.Phase[i])
		if err != nil {
			return err
		}
		if earliestPhase > attackerStateShared.CurrentPhase {
			continue
		}
		attackerStatePutKnownLocked(cc, valPrincipalState)
		attackerStatePutKnownLocked(a, valPrincipalState)
	}
	attackerStateMutex.Unlock()
	return nil
}

func attackerStateGetRead() AttackerState {
	attackerStateMutex.Lock()
	valAttackerState := attackerStateShared
	attackerStateMutex.Unlock()
	return valAttackerState
}

func attackerStateGetExhausted() bool {
	var exhausted bool
	attackerStateMutex.Lock()
	exhausted = attackerStateShared.Exhausted
	attackerStateMutex.Unlock()
	return exhausted
}

func attackerStatePutWrite(known *Value, valPrincipalState *PrincipalState) bool {
	attackerStateMutex.Lock()
	written := attackerStatePutKnownLocked(known, valPrincipalState)
	attackerStateMutex.Unlock()
	return written
}

func attackerStatePutPhaseUpdate(valKnowledgeMap *KnowledgeMap, valPrincipalState *PrincipalState, phase int) error {
	attackerStateMutex.Lock()
	attackerStateShared.CurrentPhase = phase
	attackerStateMutex.Unlock()
	err := attackerStateAbsorbPhaseValues(valKnowledgeMap, valPrincipalState)
	return err
}

func attackerStatePutExhausted() bool {
	attackerStateMutex.Lock()
	attackerStateShared.Exhausted = true
	attackerStateMutex.Unlock()
	return true
}
