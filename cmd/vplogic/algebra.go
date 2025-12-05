/* SPDX-FileCopyrightText: © 2019-2022 Nadim Kobeissi <nadim@symbolic.software>
 * SPDX-License-Identifier: GPL-3.0-only */

package vplogic

import (
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
)

const scalarExprPrefix = "scalar_"

type scalarExpr struct {
	terms    map[string]int
	constant int
}

func newScalarExprZero() scalarExpr {
	return scalarExpr{terms: map[string]int{}}
}

func (e scalarExpr) clone() scalarExpr {
	clone := newScalarExprZero()
	clone.constant = e.constant
	for k, v := range e.terms {
		clone.terms[k] = v
	}
	return clone
}

func (e scalarExpr) add(other scalarExpr) scalarExpr {
	result := e.clone()
	result.constant += other.constant
	for name, coeff := range other.terms {
		result.terms[name] += coeff
		if result.terms[name] == 0 {
			delete(result.terms, name)
		}
	}
	return result
}

func (e scalarExpr) negate() scalarExpr {
	neg := newScalarExprZero()
	neg.constant = -e.constant
	for name, coeff := range e.terms {
		neg.terms[name] = -coeff
	}
	return neg
}

func (e scalarExpr) isZero() bool {
	return e.constant == 0 && len(e.terms) == 0
}

func (e scalarExpr) normalize() scalarExpr {
	normalized := newScalarExprZero()
	normalized.constant = e.constant
	for name, coeff := range e.terms {
		if coeff != 0 {
			normalized.terms[name] = coeff
		}
	}
	return normalized
}

func (e scalarExpr) variableNames() []string {
	names := make([]string, 0, len(e.terms))
	for name := range e.terms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func scalarExprFromValue(v *Value) (scalarExpr, bool) {
	switch v.Kind {
	case typesEnumConstant:
		return scalarExprFromConstant(v.Data.(*Constant))
	case typesEnumPrimitive:
		prim := v.Data.(*Primitive)
		switch prim.ID {
		case primitiveEnumSCALARNEG:
			if len(prim.Arguments) != 1 {
				return scalarExpr{}, false
			}
			expr, ok := scalarExprFromValue(prim.Arguments[0])
			if !ok {
				return scalarExpr{}, false
			}
			return expr.negate(), true
		case primitiveEnumSCALARADD:
			if len(prim.Arguments) < 2 {
				return scalarExpr{}, false
			}
			sum := newScalarExprZero()
			for _, arg := range flattenScalarAddOperands(prim.Arguments) {
				expr, ok := scalarExprFromValue(arg)
				if !ok {
					return scalarExpr{}, false
				}
				sum = sum.add(expr)
			}
			return sum, true
		case primitiveEnumHASH, primitiveEnumPWHASH:
			return scalarExprFromHashPrimitive(prim)
		}
	}
	return scalarExpr{}, false
}

func scalarExprFromHashPrimitive(prim *Primitive) (scalarExpr, bool) {
	name := prettyPrimitiveCanonical(prim)
	expr := newScalarExprZero()
	expr.terms[name] = 1
	return expr, true
}

func scalarExprFromConstant(c *Constant) (scalarExpr, bool) {
	switch {
	case c.Name == "0":
		return newScalarExprZero(), true
	case strings.HasPrefix(c.Name, scalarExprPrefix):
		return decodeScalarConstant(c.Name)
	default:
		expr := newScalarExprZero()
		expr.terms[c.Name] = 1
		return expr, true
	}
}

func scalarExprToValue(expr scalarExpr) *Value {
	normalized := expr.normalize()
	if normalized.isZero() {
		return valueZero
	}
	name := encodeScalarExpr(normalized)
	id := valueNamesMapAdd(name)
	return &Value{
		Kind: typesEnumConstant,
		Data: &Constant{
			Name: name,
			ID:   id,
		},
	}
}

func scalarExprVariableConstantsFromValue(v *Value) []*Constant {
	expr, ok := scalarExprFromValue(v)
	if !ok {
		return []*Constant{}
	}
	names := expr.variableNames()
	constants := make([]*Constant, 0, len(names))
	for _, name := range names {
		id := valueNamesMapAdd(name)
		constants = append(constants, &Constant{Name: name, ID: id})
	}
	return constants
}

func decodeScalarConstant(name string) (scalarExpr, bool) {
	encoded := strings.TrimPrefix(name, scalarExprPrefix)
	if encoded == name {
		return scalarExpr{}, false
	}
	if encoded == "" {
		return scalarExpr{}, false
	}
	data, err := hex.DecodeString(encoded)
	if err != nil {
		return scalarExpr{}, false
	}
	expr := newScalarExprZero()
	parts := strings.Split(string(data), ";")
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return scalarExpr{}, false
		}
		coeff, err := strconv.Atoi(kv[1])
		if err != nil {
			return scalarExpr{}, false
		}
		if kv[0] == "const" {
			expr.constant = coeff
			continue
		}
		if coeff == 0 {
			continue
		}
		expr.terms[kv[0]] += coeff
		if expr.terms[kv[0]] == 0 {
			delete(expr.terms, kv[0])
		}
	}
	return expr, true
}

func encodeScalarExpr(expr scalarExpr) string {
	builder := strings.Builder{}
	if expr.constant != 0 {
		builder.WriteString("const=")
		builder.WriteString(strconv.Itoa(expr.constant))
		builder.WriteString(";")
	}
	names := expr.variableNames()
	for _, name := range names {
		builder.WriteString(name)
		builder.WriteString("=")
		builder.WriteString(strconv.Itoa(expr.terms[name]))
		builder.WriteString(";")
	}
	return scalarExprPrefix + hex.EncodeToString([]byte(builder.String()))
}

func formatScalarConstant(name string) (string, bool) {
	expr, ok := decodeScalarConstant(name)
	if !ok {
		return "", false
	}
	expr = expr.normalize()
	if expr.isZero() {
		return "0", true
	}
	parts := []string{}
	names := expr.variableNames()
	for _, n := range names {
		coeff := expr.terms[n]
		term := ""
		absCoeff := coeff
		if coeff < 0 {
			absCoeff = -coeff
		}
		switch absCoeff {
		case 0:
			continue
		case 1:
			term = n
		default:
			term = strconv.Itoa(absCoeff) + "*" + n
		}
		if coeff < 0 {
			term = "-" + term
		} else if len(parts) > 0 {
			term = "+" + term
		}
		parts = append(parts, term)
	}
	if expr.constant != 0 {
		constant := strconv.Itoa(expr.constant)
		if expr.constant > 0 && len(parts) > 0 {
			constant = "+" + constant
		}
		parts = append(parts, constant)
	}
	return strings.Join(parts, ""), true
}

func valueIsZero(v *Value) bool {
	if v.Kind != typesEnumConstant {
		return false
	}
	if v.Data.(*Constant).Name == "0" {
		return true
	}
	expr, ok := scalarExprFromValue(v)
	return ok && expr.isZero()
}

func rewriteXorPrimitive(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) < 2 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	operands := flattenXorOperands(p.Arguments)
	simplified := []*Value{}
	for _, operand := range operands {
		if valueIsZero(operand) {
			continue
		}
		removed := false
		for i, existing := range simplified {
			if valueEquivalentValues(existing, operand, true) {
				simplified = append(simplified[:i], simplified[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			simplified = append(simplified, operand)
		}
	}
	switch len(simplified) {
	case 0:
		return true, []*Value{valueZero}
	case 1:
		return true, []*Value{simplified[0]}
	default:
		rewritten := &Primitive{
			ID:        primitiveEnumXOR,
			Arguments: simplified,
			Output:    p.Output,
			Check:     p.Check,
		}
		return true, []*Value{{Kind: typesEnumPrimitive, Data: rewritten}}
	}
}

func rewriteScalarNegPrimitive(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) != 1 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	expr, ok := scalarExprFromValue(p.Arguments[0])
	if !ok {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	return true, []*Value{scalarExprToValue(expr.negate())}
}

func rewriteScalarAddPrimitive(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) < 2 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	sum := newScalarExprZero()
	for _, operand := range flattenScalarAddOperands(p.Arguments) {
		expr, ok := scalarExprFromValue(operand)
		if !ok {
			return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
		}
		sum = sum.add(expr)
	}
	sum = sum.normalize()
	return true, []*Value{scalarExprToValue(sum)}
}

func rewritePedersenCommit(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) != 2 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	vExpr, ok := scalarExprFromValue(p.Arguments[0])
	if !ok {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	rExpr, ok := scalarExprFromValue(p.Arguments[1])
	if !ok {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	vExpr = vExpr.normalize()
	rExpr = rExpr.normalize()
	if vExpr.isZero() && rExpr.isZero() {
		return true, []*Value{valueZero}
	}
	rewritten := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			scalarExprToValue(vExpr),
			scalarExprToValue(rExpr),
		},
		Output: p.Output,
		Check:  p.Check,
	}
	return true, []*Value{{Kind: typesEnumPrimitive, Data: rewritten}}
}

func rewriteNegPrimitive(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) != 1 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	arg := p.Arguments[0]
	if valueIsZero(arg) {
		return true, []*Value{valueZero}
	}
	if arg.Kind == typesEnumPrimitive {
		inner := arg.Data.(*Primitive)
		if inner.ID == primitiveEnumGROUPADD {
			_, rewritten := rewriteGroupAddPrimitive(inner)
			if len(rewritten) != 1 {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			if valueIsZero(rewritten[0]) {
				return true, []*Value{valueZero}
			}
			if rewritten[0].Kind != typesEnumPrimitive {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			inner = rewritten[0].Data.(*Primitive)
		}
		switch inner.ID {
		case primitiveEnumNEG:
			if len(inner.Arguments) != 1 {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			return true, []*Value{inner.Arguments[0]}
		case primitiveEnumPEDERSENCOMMIT:
			if len(inner.Arguments) != 2 {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			vExpr, ok := scalarExprFromValue(inner.Arguments[0])
			if !ok {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			rExpr, ok := scalarExprFromValue(inner.Arguments[1])
			if !ok {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
			neg := &Primitive{
				ID: primitiveEnumPEDERSENCOMMIT,
				Arguments: []*Value{
					scalarExprToValue(vExpr.negate()),
					scalarExprToValue(rExpr.negate()),
				},
				Output: inner.Output,
				Check:  inner.Check,
			}
			return rewritePedersenCommit(neg)
		}
	}
	return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
}

func rewriteGroupAddPrimitive(p *Primitive) (bool, []*Value) {
	if len(p.Arguments) != 2 {
		return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
	}
	operands := flattenGroupAddOperands(p.Arguments)
	sumV := newScalarExprZero()
	sumR := newScalarExprZero()
	allGenerator := true
	for _, operand := range operands {
		if valueIsZero(operand) {
			continue
		}
		generatorOnly := false
		if operand.Kind == typesEnumPrimitive {
			prim := operand.Data.(*Primitive)
			if prim.ID == primitiveEnumNEG {
				_, rewritten := rewriteNegPrimitive(prim)
				if len(rewritten) != 1 {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				operand = rewritten[0]
				if valueIsZero(operand) {
					continue
				}
				if operand.Kind != typesEnumPrimitive {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				prim = operand.Data.(*Primitive)
			}
			if prim.ID == primitiveEnumPEDERSENCOMMIT {
				if len(prim.Arguments) != 2 {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				vExpr, ok := scalarExprFromValue(prim.Arguments[0])
				if !ok {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				rExpr, ok := scalarExprFromValue(prim.Arguments[1])
				if !ok {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				sumV = sumV.add(vExpr)
				sumR = sumR.add(rExpr)
				allGenerator = false
				continue
			}
		}
		if operand.Kind == typesEnumEquation {
			eq := valueFlattenEquation(operand.Data.(*Equation))
			if len(eq.Values) == 2 && valueEquivalentValues(eq.Values[0], valueG, true) {
				vExpr, ok := scalarExprFromValue(eq.Values[1])
				if !ok {
					return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
				}
				sumV = sumV.add(vExpr)
				generatorOnly = true
			} else {
				return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
			}
		} else if operand.Kind != typesEnumPrimitive {
			return true, []*Value{{Kind: typesEnumPrimitive, Data: p}}
		}
		allGenerator = allGenerator && generatorOnly
	}
	sumV = sumV.normalize()
	sumR = sumR.normalize()
	if sumV.isZero() && sumR.isZero() {
		return true, []*Value{valueZero}
	}
	if allGenerator && sumR.isZero() {
		combined := &Value{
			Kind: typesEnumEquation,
			Data: &Equation{Values: []*Value{valueG, scalarExprToValue(sumV)}},
		}
		return true, []*Value{combined}
	}
	combined := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			scalarExprToValue(sumV),
			scalarExprToValue(sumR),
		},
		Output: p.Output,
		Check:  p.Check,
	}
	return rewritePedersenCommit(combined)
}

func flattenGroupAddOperands(args []*Value) []*Value {
	operands := []*Value{}
	for _, arg := range args {
		if arg.Kind == typesEnumPrimitive {
			prim := arg.Data.(*Primitive)
			if prim.ID == primitiveEnumGROUPADD && len(prim.Arguments) == 2 {
				operands = append(operands, flattenGroupAddOperands(prim.Arguments)...)
				continue
			}
		}
		operands = append(operands, arg)
	}
	return operands
}

func flattenScalarAddOperands(args []*Value) []*Value {
	operands := []*Value{}
	for _, arg := range args {
		if arg.Kind == typesEnumPrimitive {
			prim := arg.Data.(*Primitive)
			if prim.ID == primitiveEnumSCALARADD && len(prim.Arguments) >= 2 {
				operands = append(operands, flattenScalarAddOperands(prim.Arguments)...)
				continue
			}
		}
		operands = append(operands, arg)
	}
	return operands
}

func flattenXorOperands(args []*Value) []*Value {
	operands := []*Value{}
	for _, arg := range args {
		if arg.Kind == typesEnumPrimitive {
			prim := arg.Data.(*Primitive)
			if prim.ID == primitiveEnumXOR && len(prim.Arguments) >= 2 {
				operands = append(operands, flattenXorOperands(prim.Arguments)...)
				continue
			}
		}
		operands = append(operands, arg)
	}
	return operands
}
