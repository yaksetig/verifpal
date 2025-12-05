package vplogic

import "testing"

func TestScalarExprEncoding(t *testing.T) {
	v := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "v", ID: valueNamesMapAdd("v")}}
	expr, ok := scalarExprFromValue(v)
	if !ok {
		t.Fatalf("expected scalar expression from constant")
	}
	if expr.isZero() {
		t.Fatalf("expected non-zero scalar expression")
	}
	encoded := scalarExprToValue(expr)
	decoded, ok := scalarExprFromValue(encoded)
	if !ok {
		t.Fatalf("expected decoded scalar expression")
	}
	if !decoded.normalize().add(expr.negate()).isZero() {
		t.Fatalf("decoded expression mismatch")
	}
}

func TestPedersenCommitRewrite(t *testing.T) {
	commit := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			{Kind: typesEnumConstant, Data: &Constant{Name: "v", ID: valueNamesMapAdd("v")}},
			{Kind: typesEnumConstant, Data: &Constant{Name: "r", ID: valueNamesMapAdd("r")}},
		},
	}
	rewritten, values := rewritePedersenCommit(commit)
	if !rewritten || len(values) != 1 {
		t.Fatalf("expected pedersen commit rewrite")
	}
	result := values[0]
	if result.Kind != typesEnumPrimitive || result.Data.(*Primitive).ID != primitiveEnumPEDERSENCOMMIT {
		t.Fatalf("expected pedersen commit value")
	}
	zeroCommit := &Primitive{
		ID:        primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{scalarExprToValue(newScalarExprZero()), scalarExprToValue(newScalarExprZero())},
	}
	rewritten, values = rewritePedersenCommit(zeroCommit)
	if !rewritten || len(values) != 1 || values[0] != valueZero {
		t.Fatalf("expected pedersen commit zero rewrite")
	}
}

func TestGroupAdditionRewritesToZero(t *testing.T) {
	commit := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			{Kind: typesEnumConstant, Data: &Constant{Name: "x", ID: valueNamesMapAdd("x")}},
			{Kind: typesEnumConstant, Data: &Constant{Name: "y", ID: valueNamesMapAdd("y")}},
		},
	}
	negCommit := &Primitive{
		ID:        primitiveEnumNEG,
		Arguments: []*Value{{Kind: typesEnumPrimitive, Data: commit}},
	}
	_, negValues := rewriteNegPrimitive(negCommit)
	add := &Primitive{
		ID:        primitiveEnumGROUPADD,
		Arguments: []*Value{{Kind: typesEnumPrimitive, Data: commit}, negValues[0]},
	}
	rewritten, values := rewriteGroupAddPrimitive(add)
	if !rewritten || len(values) != 1 || values[0] != valueZero {
		t.Fatalf("expected group addition to reduce to zero")
	}
}

func TestNegDoubleNegation(t *testing.T) {
	base := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			{Kind: typesEnumConstant, Data: &Constant{Name: "a", ID: valueNamesMapAdd("a")}},
			{Kind: typesEnumConstant, Data: &Constant{Name: "b", ID: valueNamesMapAdd("b")}},
		},
	}
	neg := &Primitive{ID: primitiveEnumNEG, Arguments: []*Value{{Kind: typesEnumPrimitive, Data: base}}}
	rewritten, values := rewriteNegPrimitive(&Primitive{ID: primitiveEnumNEG, Arguments: []*Value{{Kind: typesEnumPrimitive, Data: neg}}})
	if !rewritten || len(values) != 1 {
		t.Fatalf("expected double negation rewrite")
	}
	result := values[0]
	if result.Kind != typesEnumPrimitive || result.Data.(*Primitive).ID != primitiveEnumPEDERSENCOMMIT {
		t.Fatalf("expected pedersen commit after double neg")
	}
}

func TestPreprocessLineAddition(t *testing.T) {
	line := "S1 = C + Cneg"
	processed, err := preprocessLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed != "S1 = GROUPADD(C, Cneg)" {
		t.Fatalf("unexpected preprocess result: %s", processed)
	}
	line = "Sum = PedersenCommit(a, b) + PedersenCommit(-a, -b)"
	processed, err = preprocessLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Sum = GROUPADD(PedersenCommit(a, b), PedersenCommit(SCALARNEG(a), SCALARNEG(b)))"
	if processed != expected {
		t.Fatalf("unexpected preprocess result: %s", processed)
	}
}

func TestGroupAdditionWithHashScalars(t *testing.T) {
	value := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "v", ID: valueNamesMapAdd("v")}}
	salt := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "s", ID: valueNamesMapAdd("s")}}
	block := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "n_block", ID: valueNamesMapAdd("n_block")}}
	hash := &Primitive{ID: primitiveEnumHASH, Arguments: []*Value{salt, block}}
	commit := &Primitive{ID: primitiveEnumPEDERSENCOMMIT, Arguments: []*Value{value, {Kind: typesEnumPrimitive, Data: hash}}}
	negCommit := &Primitive{
		ID: primitiveEnumPEDERSENCOMMIT,
		Arguments: []*Value{
			{Kind: typesEnumPrimitive, Data: &Primitive{ID: primitiveEnumSCALARNEG, Arguments: []*Value{value}}},
			{Kind: typesEnumPrimitive, Data: &Primitive{ID: primitiveEnumSCALARNEG, Arguments: []*Value{{Kind: typesEnumPrimitive, Data: hash}}}},
		},
	}
	groupAdd := &Primitive{ID: primitiveEnumGROUPADD, Arguments: []*Value{{Kind: typesEnumPrimitive, Data: commit}, {Kind: typesEnumPrimitive, Data: negCommit}}}
	rewritten, values := rewriteGroupAddPrimitive(groupAdd)
	if !rewritten || len(values) != 1 || values[0] != valueZero {
		t.Fatalf("expected hash-backed pedersen commits to cancel out")
	}
}

func TestXorRewriteCancellation(t *testing.T) {
	a := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "axor", ID: valueNamesMapAdd("axor")}}
	b := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "bxor", ID: valueNamesMapAdd("bxor")}}
	inner := &Primitive{ID: primitiveEnumXOR, Arguments: []*Value{a, b}}
	outer := &Primitive{ID: primitiveEnumXOR, Arguments: []*Value{{Kind: typesEnumPrimitive, Data: inner}, a}}
	rewritten, values := rewriteXorPrimitive(outer)
	if !rewritten || len(values) != 1 {
		t.Fatalf("expected xor rewrite result")
	}
	if !valueEquivalentValues(values[0], b, true) {
		t.Fatalf("expected xor rewrite to return other operand")
	}
}

func TestXorRewriteToZero(t *testing.T) {
	a := &Value{Kind: typesEnumConstant, Data: &Constant{Name: "cxor", ID: valueNamesMapAdd("cxor")}}
	xor := &Primitive{ID: primitiveEnumXOR, Arguments: []*Value{a, a}}
	rewritten, values := rewriteXorPrimitive(xor)
	if !rewritten || len(values) != 1 || values[0] != valueZero {
		t.Fatalf("expected xor of identical operands to reduce to zero")
	}
}
