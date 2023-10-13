package engine

import (
	"fmt"
	"reflect"
	"testing"
)

type TypeA struct {
	Value string
}

type TypeB struct {
	Value string
}

func (t TypeB) String() string {
	return t.Value
}

func fatal(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestValueTo(t *testing.T) {
	orig := &TypeA{Value: "A"}
	a, err := New(orig, TypeB{Value: "B"})
	fatal(t, err)

	u := a.Units()
	if len(u) != 2 {
		t.Fatal("unexpected value")
	}

	var v TypeB
	fatal(t, a.ValueTo(&v))
	if v.Value != "B" {
		t.Fatal("unexpected value")
	}

	var vv *TypeA
	fatal(t, a.ValueTo(&vv))
	if vv.Value != "A" {
		t.Fatal("unexpected value")
	}

}

func TestAssignableTo(t *testing.T) {
	a, _ := New()
	fatal(t, a.Add(
		TypeA{},
		TypeA{},
	))

	typ := reflect.TypeOf(&TypeA{})
	u := a.AssignableTo(typ)
	if len(u) != 2 {
		t.Fatalf("unexpected count: %d", len(u))
	}
}

type assembleTest struct {
	A      *TypeA
	B      []*TypeB
	I      fmt.Stringer
	hidden *TypeA
}

func TestAssemble(t *testing.T) {
	a, _ := New()
	fatal(t, a.Add(
		TypeA{},
		TypeB{Value: "B1"},
		TypeB{Value: "B2"},
	))

	v := assembleTest{}
	a.Assemble(&v)

	if v.A == nil {
		t.Fatal("unexpected nil")
	}
	if v.hidden != nil {
		t.Fatal("expected nil")
	}
	if len(v.B) != 2 {
		t.Fatal("unexpected len")
	}
	if v.I == nil {
		t.Fatal("unexpected nil")
	}
	if v.I.String() != "B1" {
		t.Fatal("unexpected value")
	}
}

type selfTypeA struct {
	TypeB *selfTypeB
}

type selfTypeB struct {
	TypeA *selfTypeA
}

func TestSelfAssemble(t *testing.T) {
	a := &selfTypeA{}
	b := &selfTypeB{}
	r, err := New(a, b)
	fatal(t, err)

	if a.TypeB != nil {
		t.Fatal("expected nil")
	}
	if b.TypeA != nil {
		t.Fatal("expected nil")
	}

	r.SelfAssemble()

	if a.TypeB != b {
		t.Fatal("expected set to b")
	}
	if b.TypeA != a {
		t.Fatal("expected set to a")
	}
}

type methodAssembleType struct {
	a *TypeA
	b *TypeB
}

func (o *methodAssembleType) Assemble(b *TypeB, a *TypeA) {
	o.a = a
	o.b = b
}

func TestMethodAssemble(t *testing.T) {
	o := &methodAssembleType{}
	a := &TypeA{}
	b := &TypeB{}
	r, err := New(o, a, b)
	fatal(t, err)

	if o.a != nil {
		t.Fatal("expected nil")
	}
	if o.b != nil {
		t.Fatal("expected nil")
	}

	r.SelfAssemble()

	if o.b != b {
		t.Fatal("expected set to b")
	}
	if o.a != a {
		t.Fatal("expected set to a")
	}
}
