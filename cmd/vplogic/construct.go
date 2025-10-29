/* SPDX-FileCopyrightText: © 2019-2022 Nadim Kobeissi <nadim@symbolic.software>
 * SPDX-License-Identifier: GPL-3.0-only */
// e7f38dcfcb1b02f4419c2e9e90efa017

package vplogic

import (
	"fmt"
)

func constructKnowledgeMap(m Model, principals []string, principalIDs []principalEnum) (*KnowledgeMap, error) {
	var err error
	valKnowledgeMap := &KnowledgeMap{
		Principals:    principals,
		PrincipalIDs:  principalIDs,
		Constants:     []*Constant{},
		Assigned:      []*Value{},
		Creator:       []principalEnum{},
		KnownBy:       [][]map[principalEnum]principalEnum{},
		DeclaredAt:    []int{},
		MaxDeclaredAt: 0,
		Phase:         [][]int{},
		MaxPhase:      0,
	}
	declaredAt := 0
	currentPhase := 0
	valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, valueG.Data.(*Constant))
	valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, valueG)
	valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, principalNamesMap["Attacker"])
	valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{})
	valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
	valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{currentPhase})
	for _, principalID := range principalIDs {
		valKnowledgeMap.KnownBy[0] = append(
			valKnowledgeMap.KnownBy[0],
			map[principalEnum]principalEnum{principalID: principalID},
		)
	}
	valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, valueNil.Data.(*Constant))
	valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, valueNil)
	valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, principalNamesMap["Attacker"])
	valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{})
	valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
	valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{currentPhase})
	for _, principalID := range principalIDs {
		valKnowledgeMap.KnownBy[1] = append(
			valKnowledgeMap.KnownBy[1],
			map[principalEnum]principalEnum{principalID: principalID},
		)
	}
	valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, valueZero.Data.(*Constant))
	valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, valueZero)
	valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, principalNamesMap["Attacker"])
	valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{})
	valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
	valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{currentPhase})
	for _, principalID := range principalIDs {
		valKnowledgeMap.KnownBy[2] = append(
			valKnowledgeMap.KnownBy[2],
			map[principalEnum]principalEnum{principalID: principalID},
		)
	}
	for _, blck := range m.Blocks {
		switch blck.Kind {
		case "principal":
			valKnowledgeMap, declaredAt, err = constructKnowledgeMapRenderPrincipal(
				valKnowledgeMap, blck, declaredAt, currentPhase,
			)
			if err != nil {
				return &KnowledgeMap{}, err
			}
		case "message":
			declaredAt = declaredAt + 1
			valKnowledgeMap.MaxDeclaredAt = declaredAt
			valKnowledgeMap, err = constructKnowledgeMapRenderMessage(
				valKnowledgeMap, blck, currentPhase,
			)
			if err != nil {
				return &KnowledgeMap{}, err
			}
		case "phase":
			currentPhase = blck.Phase.Number
		}
	}
	valKnowledgeMap.MaxPhase = currentPhase
	return valKnowledgeMap, nil
}

func constructKnowledgeMapRenderPrincipal(
	valKnowledgeMap *KnowledgeMap, blck Block, declaredAt int, currentPhase int,
) (*KnowledgeMap, int, error) {
	var err error
	for _, expr := range blck.Principal.Expressions {
		switch expr.Kind {
		case typesEnumKnows:
			valKnowledgeMap, err = constructKnowledgeMapRenderKnows(
				valKnowledgeMap, blck, declaredAt, expr,
			)
			if err != nil {
				return &KnowledgeMap{}, 0, err
			}
		case typesEnumGenerates:
			valKnowledgeMap, err = constructKnowledgeMapRenderGenerates(
				valKnowledgeMap, blck, declaredAt, expr,
			)
			if err != nil {
				return &KnowledgeMap{}, 0, err
			}
		case typesEnumAssignment:
			valKnowledgeMap, err = constructKnowledgeMapRenderAssignment(
				valKnowledgeMap, blck, declaredAt, expr,
			)
			if err != nil {
				return &KnowledgeMap{}, 0, err
			}
		case typesEnumLeaks:
			declaredAt = declaredAt + 1
			valKnowledgeMap, err = constructKnowledgeMapRenderLeaks(
				valKnowledgeMap, blck, expr, currentPhase,
			)
			if err != nil {
				return &KnowledgeMap{}, 0, err
			}
		}
	}
	return valKnowledgeMap, declaredAt, nil
}

func constructKnowledgeMapRenderKnows(
	valKnowledgeMap *KnowledgeMap, blck Block, declaredAt int, expr Expression,
) (*KnowledgeMap, error) {
	for _, c := range expr.Constants {
		i := valueGetKnowledgeMapIndexFromConstant(valKnowledgeMap, c)
		if i >= 0 {
			d1 := valKnowledgeMap.Constants[i].Declaration
			d2 := typesEnumKnows
			q1 := valKnowledgeMap.Constants[i].Qualifier
			q2 := expr.Qualifier
			fresh := valKnowledgeMap.Constants[i].Fresh
			if d1 != d2 || q1 != q2 || fresh {
				return valKnowledgeMap, fmt.Errorf(
					"constant is known more than once and in different ways (%s)",
					prettyConstant(c),
				)
			}
			valKnowledgeMap.KnownBy[i] = append(
				valKnowledgeMap.KnownBy[i],
				map[principalEnum]principalEnum{blck.Principal.ID: blck.Principal.ID},
			)
			continue
		}
		c = &Constant{
			Name:        c.Name,
			ID:          c.ID,
			Guard:       c.Guard,
			Fresh:       false,
			Leaked:      false,
			Declaration: typesEnumKnows,
			Qualifier:   expr.Qualifier,
		}
		valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, c)
		valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, &Value{
			Kind: typesEnumConstant,
			Data: c,
		})
		valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, blck.Principal.ID)
		valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{})
		valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
		valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{})
		l := len(valKnowledgeMap.Constants) - 1
		if expr.Qualifier != typesEnumPublic {
			continue
		}
		for _, principalID := range valKnowledgeMap.PrincipalIDs {
			if principalID != blck.Principal.ID {
				valKnowledgeMap.KnownBy[l] = append(
					valKnowledgeMap.KnownBy[l],
					map[principalEnum]principalEnum{principalID: principalID},
				)
			}
		}
	}
	return valKnowledgeMap, nil
}

func constructKnowledgeMapRenderGenerates(
	valKnowledgeMap *KnowledgeMap, blck Block, declaredAt int, expr Expression,
) (*KnowledgeMap, error) {
	for _, c := range expr.Constants {
		i := valueGetKnowledgeMapIndexFromConstant(valKnowledgeMap, c)
		if i >= 0 {
			return valKnowledgeMap, fmt.Errorf(
				"generated constant already exists (%s)",
				prettyConstant(c),
			)
		}
		c = &Constant{
			Name:        c.Name,
			ID:          c.ID,
			Guard:       c.Guard,
			Fresh:       true,
			Leaked:      false,
			Declaration: typesEnumGenerates,
			Qualifier:   typesEnumPrivate,
		}
		valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, c)
		valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, &Value{
			Kind: typesEnumConstant,
			Data: c,
		})
		valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, blck.Principal.ID)
		valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{{}})
		valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
		valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{})
	}
	return valKnowledgeMap, nil
}

func constructKnowledgeMapRenderAssignment(
	valKnowledgeMap *KnowledgeMap, blck Block, declaredAt int, expr Expression,
) (*KnowledgeMap, error) {
	constants, err := sanityAssignmentConstants(expr.Assigned, []*Constant{}, valKnowledgeMap)
	if err != nil {
		return &KnowledgeMap{}, err
	}
	switch expr.Assigned.Kind {
	case typesEnumPrimitive:
		err := sanityPrimitive(expr.Assigned.Data.(*Primitive), expr.Constants)
		if err != nil {
			return &KnowledgeMap{}, err
		}
	}
	for _, c := range constants {
		i := valueGetKnowledgeMapIndexFromConstant(valKnowledgeMap, c)
		if i < 0 {
			return valKnowledgeMap, fmt.Errorf(
				"constant does not exist (%s)",
				prettyConstant(c),
			)
		}
		knows := valKnowledgeMap.Creator[i] == blck.Principal.ID
		for _, m := range valKnowledgeMap.KnownBy[i] {
			if _, ok := m[blck.Principal.ID]; ok {
				knows = true
				break
			}
		}
		if !knows {
			return valKnowledgeMap, fmt.Errorf(
				"%s is using constant (%s) despite not knowing it",
				blck.Principal.Name,
				prettyConstant(c),
			)
		}
	}
	for i, c := range expr.Constants {
		ii := valueGetKnowledgeMapIndexFromConstant(valKnowledgeMap, c)
		if ii >= 0 {
			return valKnowledgeMap, fmt.Errorf(
				"constant assigned twice (%s)",
				prettyConstant(c),
			)
		}
		c = &Constant{
			Name:        c.Name,
			ID:          c.ID,
			Guard:       c.Guard,
			Fresh:       false,
			Leaked:      false,
			Declaration: typesEnumAssignment,
			Qualifier:   typesEnumPrivate,
		}
		a := valueDeepCopy(expr.Assigned)
		switch a.Kind {
		case typesEnumPrimitive:
			a.Data.(*Primitive).Output = i
		}
		valKnowledgeMap.Constants = append(valKnowledgeMap.Constants, c)
		valKnowledgeMap.Assigned = append(valKnowledgeMap.Assigned, &a)
		valKnowledgeMap.Creator = append(valKnowledgeMap.Creator, blck.Principal.ID)
		valKnowledgeMap.KnownBy = append(valKnowledgeMap.KnownBy, []map[principalEnum]principalEnum{{}})
		valKnowledgeMap.DeclaredAt = append(valKnowledgeMap.DeclaredAt, declaredAt)
		valKnowledgeMap.Phase = append(valKnowledgeMap.Phase, []int{})
	}
	return valKnowledgeMap, nil
}

func constructKnowledgeMapRenderLeaks(
	valKnowledgeMap *KnowledgeMap, blck Block, expr Expression, currentPhase int,
) (*KnowledgeMap, error) {
	for _, c := range expr.Constants {
		i := valueGetKnowledgeMapIndexFromConstant(
			valKnowledgeMap, c,
		)
		if i < 0 {
			return valKnowledgeMap, fmt.Errorf(
				"leaked constant does not exist (%s)",
				prettyConstant(c),
			)
		}
		known := valKnowledgeMap.Creator[i] == blck.Principal.ID
		for _, m := range valKnowledgeMap.KnownBy[i] {
			if _, ok := m[blck.Principal.ID]; ok {
				known = true
				break
			}
		}
		if !known {
			return valKnowledgeMap, fmt.Errorf(
				"%s leaks a constant that they do not know (%s)",
				blck.Principal.Name, prettyConstant(c),
			)
		}
		valKnowledgeMap.Constants[i].Leaked = true
		valKnowledgeMap.Phase[i], _ = appendUniqueInt(
			valKnowledgeMap.Phase[i], currentPhase,
		)
	}
	return valKnowledgeMap, nil
}

func constructKnowledgeMapRenderMessage(
	valKnowledgeMap *KnowledgeMap, blck Block, currentPhase int,
) (*KnowledgeMap, error) {
	for _, c := range blck.Message.Constants {
		i := valueGetKnowledgeMapIndexFromConstant(valKnowledgeMap, c)
		if i < 0 {
			return valKnowledgeMap, fmt.Errorf(
				"%s sends unknown constant to %s (%s)",
				principalGetNameFromID(blck.Message.Sender),
				principalGetNameFromID(blck.Message.Recipient),
				prettyConstant(c),
			)
		}
		c = valKnowledgeMap.Constants[i]
		senderKnows := false
		recipientKnows := false
		if valKnowledgeMap.Creator[i] == blck.Message.Sender {
			senderKnows = true
		}
		for _, m := range valKnowledgeMap.KnownBy[i] {
			if _, ok := m[blck.Message.Sender]; ok {
				senderKnows = true
			}
		}
		if valKnowledgeMap.Creator[i] == blck.Message.Recipient {
			recipientKnows = true
		}
		for _, m := range valKnowledgeMap.KnownBy[i] {
			if _, ok := m[blck.Message.Recipient]; ok {
				recipientKnows = true
			}
		}
		switch {
		case !senderKnows:
			return valKnowledgeMap, fmt.Errorf(
				"%s is sending constant (%s) despite not knowing it",
				principalGetNameFromID(blck.Message.Sender),
				prettyConstant(c),
			)
		case recipientKnows:
			return valKnowledgeMap, fmt.Errorf(
				"%s is receiving constant (%s) despite already knowing it",
				principalGetNameFromID(blck.Message.Recipient),
				prettyConstant(c),
			)
		}
		valKnowledgeMap.KnownBy[i] = append(
			valKnowledgeMap.KnownBy[i], map[principalEnum]principalEnum{
				blck.Message.Recipient: blck.Message.Sender,
			},
		)
		valKnowledgeMap.Phase[i], _ = appendUniqueInt(
			valKnowledgeMap.Phase[i], currentPhase,
		)
	}
	return valKnowledgeMap, nil
}

func constructPrincipalStates(m Model, valKnowledgeMap *KnowledgeMap) []*PrincipalState {
	valPrincipalStates := []*PrincipalState{}
	for p := range valKnowledgeMap.Principals {
		valPrincipalState := &PrincipalState{
			Name:          valKnowledgeMap.Principals[p],
			ID:            valKnowledgeMap.PrincipalIDs[p],
			Constants:     []*Constant{},
			Assigned:      []*Value{},
			Guard:         []bool{},
			Known:         []bool{},
			Wire:          [][]principalEnum{},
			KnownBy:       [][]map[principalEnum]principalEnum{},
			DeclaredAt:    []int{},
			MaxDeclaredAt: 0,
			Creator:       []principalEnum{},
			Sender:        []principalEnum{},
			Rewritten:     []bool{},
			BeforeRewrite: []*Value{},
			Mutated:       []bool{},
			MutatableTo:   [][]principalEnum{},
			BeforeMutate:  []*Value{},
			Phase:         [][]int{},
		}
		for i, c := range valKnowledgeMap.Constants {
			wire := []principalEnum{}
			guard := false
			mutatableTo := []principalEnum{}
			knows := false
			sender := valKnowledgeMap.Creator[i]
			assigned := valKnowledgeMap.Assigned[i]
			if valKnowledgeMap.Creator[i] == valKnowledgeMap.PrincipalIDs[p] {
				knows = true
			}
			for _, m := range valKnowledgeMap.KnownBy[i] {
				if precedingSender, ok := m[valKnowledgeMap.PrincipalIDs[p]]; ok {
					sender = precedingSender
					knows = true
					break
				}
			}
			for _, blck := range m.Blocks {
				wire, guard, mutatableTo = constructPrincipalStatesGetValueMutatability(
					c, blck, valKnowledgeMap.PrincipalIDs[p], valKnowledgeMap.Creator[i],
					wire, guard, mutatableTo,
				)
			}
			valPrincipalState.Constants = append(valPrincipalState.Constants, c)
			valPrincipalState.Assigned = append(valPrincipalState.Assigned, assigned)
			valPrincipalState.Guard = append(valPrincipalState.Guard, guard)
			valPrincipalState.Known = append(valPrincipalState.Known, knows)
			valPrincipalState.Wire = append(valPrincipalState.Wire, wire)
			valPrincipalState.KnownBy = append(valPrincipalState.KnownBy, valKnowledgeMap.KnownBy[i])
			valPrincipalState.DeclaredAt = append(valPrincipalState.DeclaredAt, valKnowledgeMap.DeclaredAt[i])
			valPrincipalState.MaxDeclaredAt = valKnowledgeMap.MaxDeclaredAt
			valPrincipalState.Creator = append(valPrincipalState.Creator, valKnowledgeMap.Creator[i])
			valPrincipalState.Sender = append(valPrincipalState.Sender, sender)
			valPrincipalState.Rewritten = append(valPrincipalState.Rewritten, false)
			valPrincipalState.BeforeRewrite = append(valPrincipalState.BeforeRewrite, assigned)
			valPrincipalState.Mutated = append(valPrincipalState.Mutated, false)
			valPrincipalState.MutatableTo = append(valPrincipalState.MutatableTo, mutatableTo)
			valPrincipalState.BeforeMutate = append(valPrincipalState.BeforeMutate, assigned)
			valPrincipalState.Phase = append(valPrincipalState.Phase, valKnowledgeMap.Phase[i])
		}
		valPrincipalStates = append(valPrincipalStates, valPrincipalState)
	}
	return valPrincipalStates
}

func constructPrincipalStatesGetValueMutatability(
	c *Constant, blck Block, principalID principalEnum, creator principalEnum,
	wire []principalEnum, guard bool, mutatableTo []principalEnum,
) ([]principalEnum, bool, []principalEnum) {
	switch blck.Kind {
	case "message":
		ir := (blck.Message.Recipient == principalID)
		ic := (creator == principalID)
		for _, cc := range blck.Message.Constants {
			if c.ID != cc.ID {
				continue
			}
			wire, _ = appendUniquePrincipalEnum(wire, blck.Message.Recipient)
			if !guard {
				guard = cc.Guard && (ir || ic)
			}
			if !cc.Guard {
				mutatableTo, _ = appendUniquePrincipalEnum(
					mutatableTo, blck.Message.Recipient,
				)
			}
		}
	}
	return wire, guard, mutatableTo
}

func constructPrincipalStateClone(valPrincipalState *PrincipalState, purify bool) *PrincipalState {
	valPrincipalStateClone := PrincipalState{
		Name:          valPrincipalState.Name,
		ID:            valPrincipalState.ID,
		Constants:     make([]*Constant, len(valPrincipalState.Constants)),
		Assigned:      make([]*Value, len(valPrincipalState.Assigned)),
		Guard:         make([]bool, len(valPrincipalState.Guard)),
		Known:         make([]bool, len(valPrincipalState.Known)),
		Wire:          make([][]principalEnum, len(valPrincipalState.Wire)),
		KnownBy:       make([][]map[principalEnum]principalEnum, len(valPrincipalState.KnownBy)),
		DeclaredAt:    make([]int, len(valPrincipalState.DeclaredAt)),
		MaxDeclaredAt: valPrincipalState.MaxDeclaredAt,
		Creator:       make([]principalEnum, len(valPrincipalState.Creator)),
		Sender:        make([]principalEnum, len(valPrincipalState.Sender)),
		Rewritten:     make([]bool, len(valPrincipalState.Rewritten)),
		BeforeRewrite: make([]*Value, len(valPrincipalState.BeforeRewrite)),
		Mutated:       make([]bool, len(valPrincipalState.Mutated)),
		MutatableTo:   make([][]principalEnum, len(valPrincipalState.MutatableTo)),
		BeforeMutate:  make([]*Value, len(valPrincipalState.BeforeMutate)),
		Phase:         make([][]int, len(valPrincipalState.Phase)),
	}
	copy(valPrincipalStateClone.Constants, valPrincipalState.Constants)
	if purify {
		copy(valPrincipalStateClone.Assigned, valPrincipalState.BeforeMutate)
	} else {
		copy(valPrincipalStateClone.Assigned, valPrincipalState.Assigned)
	}
	copy(valPrincipalStateClone.Guard, valPrincipalState.Guard)
	copy(valPrincipalStateClone.Known, valPrincipalState.Known)
	copy(valPrincipalStateClone.Wire, valPrincipalState.Wire)
	copy(valPrincipalStateClone.KnownBy, valPrincipalState.KnownBy)
	copy(valPrincipalStateClone.DeclaredAt, valPrincipalState.DeclaredAt)
	copy(valPrincipalStateClone.Creator, valPrincipalState.Creator)
	copy(valPrincipalStateClone.Sender, valPrincipalState.Sender)
	copy(valPrincipalStateClone.Rewritten, valPrincipalState.Rewritten)
	if purify {
		copy(valPrincipalStateClone.BeforeRewrite, valPrincipalState.BeforeMutate)
	} else {
		copy(valPrincipalStateClone.BeforeRewrite, valPrincipalState.BeforeRewrite)
	}
	copy(valPrincipalStateClone.Mutated, valPrincipalState.Mutated)
	copy(valPrincipalStateClone.MutatableTo, valPrincipalState.MutatableTo)
	copy(valPrincipalStateClone.BeforeMutate, valPrincipalState.BeforeMutate)
	copy(valPrincipalStateClone.Phase, valPrincipalState.Phase)
	return &valPrincipalStateClone
}
