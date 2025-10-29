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
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "v", ID: valueNamesMapAdd("v")}},
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "r", ID: valueNamesMapAdd("r")}},
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
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "x", ID: valueNamesMapAdd("x")}},
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "y", ID: valueNamesMapAdd("y")}},
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
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "a", ID: valueNamesMapAdd("a")}},
			&Value{Kind: typesEnumConstant, Data: &Constant{Name: "b", ID: valueNamesMapAdd("b")}},
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
