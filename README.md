# Lidi
>"My Lite implementation Dependency Injection for Golang"

# Overview
Lidi is a small golang library for DI:
- Constructor injection
- Setter injection
- Field injection

```bash
go get -u github.com/ekprog/lidi
```


```go
package main 

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

func main() {
	c := lidi.NewLidi(lidi.Settings{
		InvokeErrCheck: true, // Check return value
	})

	someVar := "Hello world!!!"
	s1 := &Service1{}
	s2 := &Service2{}
	s3 := &Service3{}

	// Provide named var
	if err := c.Provide(someVar, lidi.Name("hello_var")); err != nil {
		panic(err)
	}

	// Provide Service1
	if err := c.Provide(s1); err != nil {
		panic(err)
	}

	// Provide Service2
	if err := c.Provide(s2); err != nil {
		panic(err)
	}

	// Provide Service3
	if err := c.Provide(s3); err != nil {
		panic(err)
	}

	// Invoke
	if err := c.InvokeFunction(Invoke); err != nil {
		panic(err)
	}
}
```

# License
Lidi is distributed under MIT.
