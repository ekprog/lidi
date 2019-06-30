package lidi

import (
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"
)

//
// Node is simple holder for instance
//
type node struct {
	id    string
	value reflect.Value
}

type nodes map[string]node

func newNode(id string, value reflect.Value) node {
	return node{
		id:    id,
		value: value,
	}
}

func newNodes() nodes {
	return make(map[string]node)
}

// Helpers

func (c *Lidi) elem(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		return t.Elem()
	}
	return t
}

//
// Lidi - main container and DI provider
//
const (
	lidi_tag   = "lidi"
	inject_tag = "inject"
	name_tag   = "name"
)

type param_list []string

type dependency_list []reflect.Value

var tag_re = regexp.MustCompile("(.*?)\\((.*?)\\)")

type tagValue struct {
	hasInjector bool
	hasName     bool
	injector    string
	name        string
}

type Option struct {
	name string
}

func Name(name string) Option {
	return Option{name:name}
}

func (o *Option) apply(d *dependencyOptions) {
	d.name = o.name
}
type dependencyOptions struct {
	name string
}

func (obj *dependencyOptions) apply(option Option) {
	option.apply(obj)
}

func newTagValue(tag string) (tagValue, error) {
	tag_v := tagValue{}
	err := errors.Errorf("lidi warning: incorrect tag_v format: %s\n", tag)
	fs := strings.Split(tag, ",")
	for i := range fs {
		fs[i] = strings.TrimSpace(fs[i])
		match := tag_re.FindStringSubmatch(fs[i])
		if len(match) < 3 {
			return tagValue{}, err
		}
		fname, fparam := match[1], match[2]
		if fname == "" && fparam == "" {
			return tagValue{}, err
		}

		switch fname {
		case inject_tag:
			tag_v.injector = fparam
			tag_v.hasInjector = true
		case name_tag:
			tag_v.name = fparam
			tag_v.hasName = true
		default:
			return tagValue{}, err
		}
	}
	return tag_v, nil
}

type Settings struct {
	InvokeErrCheck bool
}

type Lidi struct {
	settings Settings
	typed    nodes
	key_p    string
}

func NewLidi(settings Settings) *Lidi {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyzåäö")
	var b strings.Builder
	b.WriteRune('_')
	for i := 0; i < 3; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}

	return &Lidi{
		settings: settings,
		typed:    newNodes(),
		key_p:    b.String(),
	}
}

func (c *Lidi) makeId(k string) string {
	return k + c.key_p
}

func (c *Lidi) setValue(k string, v reflect.Value) {
	c.typed[c.makeId(k)] = newNode(c.makeId(k), v)
}

func (c *Lidi) getValue(k string) (reflect.Value, error) {
	if node, ok := c.typed[c.makeId(k)]; ok {
		return node.value, nil
	} else {
		return reflect.Value{}, errors.Errorf("lidi: dependency '%s' not found", k)
	}
}

func (c *Lidi) isExists(k string) bool {
	_, ok := c.typed[c.makeId(k)]
	return ok
}

// Provide singleton instances.
// Each instance unique by type.
// If instance is service (struct that require other dependencies),
// all inner dependencies will be inject
func (c *Lidi) Provide(d interface{}, options ...Option) error {

	optionsProvider := &dependencyOptions{}
	for _, option := range options {
		optionsProvider.apply(option)
	}

	d_type := reflect.TypeOf(d)
	d_value := reflect.ValueOf(d)
	type_id := d_type.String()

	if optionsProvider.name == "" {
		optionsProvider.name = type_id
	}

	if c.isExists(type_id) {
		return errors.Errorf("lidi: dependency '%s' already exists", type_id)
	}

	// Resolve dependencies for services
	if c.elem(d_type).Kind() == reflect.Struct {
		err := c.resolveService(d_value)
		if err != nil {
			return err
		}
	}

	// Add dependency in container
	c.setValue(optionsProvider.name, d_value)

	return nil
}

func (c *Lidi) resolveService(d_value reflect.Value) error {
	d_elem := c.elem(d_value.Type())
	for i := 0; i < d_elem.NumField(); i++ {
		f_t := d_elem.Field(i)
		tag := f_t.Tag.Get(lidi_tag)
		if tag == "" {
			continue
		}

		f_v := reflect.Indirect(d_value).Field(i)
		if !f_v.IsValid() {
			return errors.Errorf("lidi: got invalid field '%s'", f_v.Type())
		}

		tag_v, err := newTagValue(tag)
		if err != nil {
			return err
		}

		if !tag_v.hasName {
			tag_v.name = f_v.Type().String()
		}

		if !tag_v.hasInjector {
			log.Printf("lidi warning: field has not 'inject' tagValue: %s\n", d_value.Type().String())
			continue
		}

		if tag_v.injector == "" { //forward
			err := c.injectForward(f_v, tag_v.name)
			if err != nil {
				return err
			}
		} else { // setter
			method := d_value.MethodByName(tag_v.injector)
			if !method.IsValid() {
				return errors.Errorf("lidi: setter method '%s' not found", tag_v.injector)
			}
			method_t := method.Type()
			if method_t.NumIn() != 1 {
				return errors.Errorf("lidi: setter method '%s' cannot take more than one param",
					tag_v.injector)
			}
			p, err := c.getValue(tag_v.name)
			if err != nil {
				return err
			}
			if p.Type() != method_t.In(0) {
				return errors.Errorf("lidi: setter method '%s' cannot take param with type '%s'",
					tag_v.injector, p.Type())
			}
			err = c.invokeFunction(method, []reflect.Value{p})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Resolve dependencies for one field in service
func (c *Lidi) injectForward(f_v reflect.Value, name string) error {
	if !f_v.CanSet() {
		return errors.Errorf("lidi: cannot inject service in unexported field '%s'", name)
	}
	k := name
	if name == "" {
		k = f_v.Type().String()
	}
	d, err := c.getValue(k)
	if err != nil {
		return err
	} else {
		f_v.Set(d)
		return nil
	}
}

// Call any function with auto dependency injection by type.
// For named dependencies use specific setter.
// e.g:
// ExportedVar   int `lidi:"named(my_int_var)"`
// unexportedVar int `lidi:"named(my_int_var, SomeExportedSetterName)"`
//
func (c *Lidi) InvokeFunction(function interface{}) error {
	func_t := reflect.TypeOf(function)
	func_v := reflect.ValueOf(function)

	pl := c.buildParamsKeys(func_t)
	args, err := c.buildParamValues(pl)
	if err != nil {
		return err
	}
	return c.invokeFunction(func_v, args)
}

func (c *Lidi) invokeFunction(func_v reflect.Value, args []reflect.Value) error {
	func_t := func_v.Type()
	if func_t == nil {
		return errors.New("lidi: can't invoke an nil func")
	}
	if func_t.Kind() != reflect.Func {
		return errors.Errorf("lidi: can't invoke non-function (type %value)", func_t)
	}
	// Call function
	rp := func_v.Call(args)
	if !c.settings.InvokeErrCheck {
		return nil
	}

	// Check if returns error
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	for i := 0; i < func_t.NumOut(); i++ {
		if func_t.Out(i).Implements(errorInterface) {
			r_i, ok := rp[i].Interface().(error)
			if !ok {
				continue
			}
			if r_i != nil {
				return r_i
			}
		}
	}
	return nil
}

// Building param list with required dependency types
func (c *Lidi) buildParamsKeys(func_t reflect.Type) param_list {
	numArgs := func_t.NumIn()
	if func_t.IsVariadic() {
		numArgs--
	}
	if numArgs == 0 {
		return param_list{}
	}
	var pl param_list
	for i := 0; i < numArgs; i++ {
		p_type := func_t.In(i)
		pl = append(pl, p_type.String())
	}
	return pl
}

// Building param list with required dependency values
func (c *Lidi) buildParamValues(pl param_list) (dependency_list, error) {
	var args dependency_list
	for i := range pl {
		d, err := c.getValue(pl[i])
		if err != nil {
			return nil, err
		} else {
			args = append(args, d)
		}
	}
	if len(args) == 0 {
		return dependency_list{}, nil
	}
	return args, nil
}