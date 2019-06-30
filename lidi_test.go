package lidi

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func Test_IntProvide(t *testing.T) {
	c := NewLidi(Settings{})
	a := rand.Int()
	err := c.Provide(a)
	if err != nil {
		t.Error(err)
	}
	err = c.InvokeFunction(func(b int) {
		if a != b {
			t.Fatal("Not equal")
		}
	})
	if err != nil {
		t.Error(err)
	}
}

func Test_IntPtrProvide(t *testing.T) {
	c := NewLidi(Settings{})
	a := rand.Int()
	err := c.Provide(&a)
	if err != nil {
		t.Error(err)
	}
	err = c.InvokeFunction(func(b *int) {
		if &a != b {
			t.Fatal("Not equal")
		}
	})
	if err != nil {
		t.Error(err)
	}

	c = NewLidi(Settings{})
	a = rand.Int()
	err = c.Provide(&a)
	if err != nil {
		t.Error(err)
	}
	err = c.InvokeFunction(func(b int) {
		t.Error(err)
	})
	if err != nil {
		if err.Error() != fmt.Sprintf("lidi: dependency '%s' not found", reflect.TypeOf(0).String()) {
			t.Fatal("Not equal")
		}
	}
}

func Test_ServicePtrProvide(t *testing.T) {
	c := NewLidi(Settings{})

	type S1 struct {
		test string
	}
	s1 := &S1{"awesome"}

	err := c.Provide(s1)
	if err != nil {
		t.Fatal(err)
	}
	err = c.InvokeFunction(func(s *S1) {
		if s1 != s {
			t.Fatal("Not equal")
		}
		if s1.test != s.test {
			t.Fatal("Not equal")
		}
	})
	if err != nil {
		t.Error(err)
	}
	err = c.InvokeFunction(func(s S1) {
		t.Error(errors.New("Incompatible types passed"))
	})
	if err != nil {
		if err.Error() != fmt.Sprintf("lidi: dependency '%s' not found", reflect.TypeOf(S1{}).String()) {
			t.Fatal("Not equal")
		}
	}
}

func Test_ServiceProvide(t *testing.T) {
	c := NewLidi(Settings{})

	type S1 struct {
		test string
	}
	s1 := S1{"awesome"}

	err := c.Provide(s1)
	if err != nil {
		t.Error(err)
	}

	s1.test = "changed"
	err = c.InvokeFunction(func(s S1) {
		if s1.test != "changed" {
			t.Fatal("Not equal")
		}
		if s.test != "awesome" {
			t.Fatal("Not equal")
		}
		if s1.test != "changed" {
			t.Fatal("Equal")
		}
		if &s1 == &s {
			t.Fatal("Equal")
		}
		if s1 == s {
			t.Fatal("Equal")
		}
	})
	if err != nil {
		t.Error(err)
	}
}

type S3 struct {
	Service1 string `lidi:"inject()"`
	Service2 int    `lidi:"inject(Setter)"`
}

func (s *S3) Setter(s1 int) {
	s.Service2 = s1
}

func Test_ServiceFieldProvide(t *testing.T) {
	c := NewLidi(Settings{})

	s1 := "awesome"
	s2 := 15
	s3 := &S3{}

	if err := c.Provide(s1); err != nil {
		t.Fatal(err)
	}
	if err := c.Provide(s2); err != nil {
		t.Fatal(err)
	}
	if err := c.Provide(s3); err != nil {
		t.Fatal(err)
	}

	err := c.InvokeFunction(func(s *S3) {
		if s.Service1 != "awesome" {
			t.Fatal("Not Equal")
		}
		if s.Service2 != 15 {
			t.Fatal("Not Equal")
		}
	})
	if err != nil {
		t.Error(err)
	}
}

func Test_DependencyExists(t *testing.T) {
	c := NewLidi(Settings{})

	s1 := "awesome"
	s2 := "other"

	if err := c.Provide(s1); err != nil {
		t.Fatal(err)
	}
	if err := c.Provide(s1); err != nil {
		if err.Error() != fmt.Sprintf("lidi: dependency '%s' already exists", reflect.TypeOf(s2).String()) {
			t.Fatal("Not Equal")
		}
	} else {
		t.Fatal("the same types in container")
	}
}

func Test_Unexported(t *testing.T) {
	c1 := NewLidi(Settings{})

	type A struct{}
	type B struct {
		a *A `lidi:"inject()"`
	}
	if err := c1.Provide(&A{}); err != nil {
		t.Fatal(err)
	}
	if err := c1.Provide(&B{}); err != nil {
		if err.Error() != fmt.Sprintf("lidi: cannot inject service in unexported field '%s'", reflect.TypeOf(&A{}).String()) {
			t.Fatal("Not Equal")
		}
	} else {
		t.Fatal(err)
	}
}

func Test_SetterNotFound(t *testing.T) {
	c1 := NewLidi(Settings{})

	type A struct{}
	type B struct {
		a *A `lidi:"inject(MySetter)"`
	}
	if err := c1.Provide(&A{}); err != nil {
		t.Fatal(err)
	}
	if err := c1.Provide(&B{}); err != nil {
		if err.Error() != "lidi: setter method 'MySetter' not found" {
			t.Fatal("Not Equal")
		}
	} else {
		t.Fatal(err)
	}
}

// TEST SETTERS

type A struct {
	test string
}
type B struct {
	a *A `lidi:"inject(MySetter)"`
}

func (b *B) MySetter(a *A) {
	b.a = a
}

func Test_Setter(t *testing.T) {
	c1 := NewLidi(Settings{})

	a := &A{"awesome"}
	b := &B{}

	if err := c1.Provide(a); err != nil {
		t.Fatal(err)
	}
	if err := c1.Provide(b); err != nil {
		t.Fatal(err)
	}
	if err := c1.InvokeFunction(func(b *B) {
		if b.a.test != "awesome" {
			t.Fatal("Not Equal")
		}
	}); err != nil {
		t.Fatal(err)
	}
}

type OneInjector struct {
	a int `lidi:"inject(Injecter)"`
}

func (obj *OneInjector) Injecter(a int, b string) {
}

func Test_AnyParamsInSetter(t *testing.T) {
	c1 := NewLidi(Settings{})

	if err := c1.Provide(15); err != nil {
		t.Fatal(err)
	}
	if err := c1.Provide("awesome"); err != nil {
		t.Fatal(err)
	}
	if err := c1.Provide(&OneInjector{}); err != nil {
		if err.Error() != "lidi: setter method 'Injecter' cannot take more than one param" {
			t.Fatal("Not Equal")
		}
	} else {
		t.Fatal(err)
	}
}

type A1 struct {
	data1 int
	data2 string
}

type A2 struct {
	Service1 *A1 `lidi:"inject(),name(name1)"`
}

func Test_Params(t *testing.T) {
	c1 := NewLidi(Settings{
		InvokeErrCheck: true,
	})

	if err := c1.Provide(&A1{}, Name("name1")); err != nil {
		t.Fatal(err)
	}

	a2 := &A2{}
	if err := c1.Provide(a2); err != nil {
		t.Fatal(err)
	}
	if a2.Service1 == nil {
		t.Fatal("Nil")
	}
}

type ErrCheck struct {
	testData int `lidi:"inject(Inject)"`
}

func (obj *ErrCheck) Inject(val int) error {
	obj.testData = val
	return errors.New("some error")
}

func Test_ErrCheck(t *testing.T) {
	c1 := NewLidi(Settings{
		InvokeErrCheck: true,
	})

	if err := c1.Provide(15); err != nil {
		t.Fatal(err)
	}

	v := &ErrCheck{}
	if err := c1.Provide(v); err != nil {
		if err.Error() != "some error" {
			t.Fatal("Not Equal")
		}
	} else {
		t.Fatal(err)
	}
}
