# Lidi
>"My Lite implementation Dependency Injection for Golang"

```bash
go get -u gopkg.in/ekprog/lidi
```

# Overview
Lidi is a small golang library for DI:
- Constructor injection
- Setter injection
- Field injection


```
package main 

import ...

type Service1 struct {
	FieldInject string `lidi:"inject(), name(hello_var)"` //Inject by field with name
}

type Service2 struct {
	Service1 *Service1 `lidi:"inject()"` //Inject by Field
}

type Service3 struct {
	service2 *Service2 `lidi:"inject(SetService)"` //Inject by setter
}

func (s *Service3) SetService(service2 *Service2) error {
	s.service2 = service2
	return nil //or error
}

func Invoke(s3 *Service3) {
	fmt.Println(s3.service2.Service1.FieldInject)
}

func Test(t *testing.T) {
	lidi := NewLidi(Settings{
		InvokeErrCheck: true, // Check return value
	})

	someVar := "Hello world!!!"
	s1 := &Service1{}
	s2 := &Service2{}
	s3 := &Service3{}

	// Provide named var
	if err := lidi.Provide(someVar, Name("hello_var")); err != nil {
		panic(err)
	}

	// Provide Service1
	if err := lidi.Provide(s1); err != nil {
		panic(err)
	}

	// Provide Service2
	if err := lidi.Provide(s2); err != nil {
		panic(err)
	}

	// Provide Service3
	if err := lidi.Provide(s3); err != nil {
		panic(err)
	}

	// Invoke
	if err := lidi.InvokeFunction(Invoke); err != nil {
		panic(err)
	}
}
```

# License
Telebot is distributed under MIT.