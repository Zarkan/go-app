package markup

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Component is the interface that describes a component.
// Should be implemented on a non empty struct pointer.
type Component interface {
	// Render should return a string describing the component with HTML5
	// standard.
	// It supports standard Go html/template API.
	// Pipeline is based on the component struct.
	// See https://golang.org/pkg/text/template and
	// https://golang.org/pkg/html/template for template usage.
	Render() string
}

// Mounter is the interface that wraps OnMount method.
// OnMount si called when a component is mounted.
type Mounter interface {
	OnMount()
}

// Dismounter is the interface that wraps OnDismount method.
// OnDismount si called when a component is dismounted.
type Dismounter interface {
	OnDismount()
}

// Navigator is the interface that wraps OnNavigate method.
// OnNavigate is called when a component is navigated to.
type Navigator interface {
	OnNavigate(u url.URL)
}

// Mapper is the interface that wraps FuncMaps method.
type Mapper interface {
	// Allows to add custom functions to the template used to render the
	// component.
	//
	// Funcs named json and time are reserved. They handle json conversion and
	// time format.
	// They can't be overloaded.
	// See https://golang.org/pkg/text/template/#Template.Funcs for more details.
	FuncMaps() template.FuncMap
}

// ZeroCompo is the type to redefine when writing an empty component.
// Every instances of an empty struct is given the same memory address, which
// causes problem for indexing components.
// ZeroCompo have a placeholder field to avoid that.
type ZeroCompo struct {
	placeholder byte
}

func mapComponentFields(c Component, attrs AttrMap) error {
	if len(attrs) == 0 {
		return nil
	}

	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i, numField := 0, t.NumField(); i < numField; i++ {
		f := v.Field(i)
		finfo := t.Field(i)

		if finfo.Anonymous {
			continue
		}

		if len(finfo.PkgPath) != 0 {
			continue
		}

		key := strings.ToLower(finfo.Name)
		val, ok := attrs[key]
		if !ok {
			if f.Kind() == reflect.Bool {
				f.SetBool(false)
			}
			continue
		}

		if err := mapComponentField(f, val); err != nil {
			return errors.Wrapf(err, `fail to map %s="%s" to %T.%s`, key, val, c, finfo.Name)
		}
	}
	return nil
}

func mapComponentField(f reflect.Value, v string) error {
	switch f.Kind() {
	case reflect.String:
		f.SetString(v)

	case reflect.Bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		f.SetBool(b)

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		n, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return err
		}
		f.SetInt(n)

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uintptr:
		n, err := strconv.ParseUint(v, 0, 64)
		if err != nil {
			return err
		}
		f.SetUint(n)

	case reflect.Float64, reflect.Float32:
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		f.SetFloat(n)

	default:
		addr := f.Addr()
		i := addr.Interface()
		if err := json.Unmarshal([]byte(v), i); err != nil {
			return err
		}
	}
	return nil
}

func decodeComponent(c Component, root *Tag) error {
	var funcMap template.FuncMap
	if mapper, ok := c.(Mapper); ok {
		funcMap = mapper.FuncMaps()
	}
	if len(funcMap) == 0 {
		funcMap = make(template.FuncMap, 3)
	}
	funcMap["raw"] = rawHTML
	funcMap["json"] = convertToJSON
	funcMap["time"] = formatTime

	r := c.Render()
	tmpl := template.Must(template.New("").Funcs(funcMap).Parse(r))

	b := bytes.Buffer{}
	if err := tmpl.Execute(&b, c); err != nil {
		return errors.Wrapf(err, "fail to decode %T", c)
	}

	dec := NewTagDecoder(&b)
	if err := dec.Decode(root); err != nil {
		return errors.Wrapf(err, "fail to decode %T", c)
	}
	return nil
}

func rawHTML(s string) template.HTML {
	return template.HTML(s)
}

func convertToJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func formatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// CallOrAssign call the method n with jval mapped as the first arg or assign
// jval to the field named n.
// Methods must take 0 or 1 arg and no return values.
// Methods and and fields must be exported.
func CallOrAssign(c Component, n string, jval string) error {
	structval := reflect.ValueOf(c)

	if m := structval.MethodByName(n); m.IsValid() {
		return callComponentMethod(m, jval)
	}

	structval = structval.Elem()

	if f := structval.FieldByName(n); f.IsValid() {
		return assignComponentField(f, jval)
	}

	return errors.Errorf("no method or field named %v", n)
}

func callComponentMethod(m reflect.Value, jval string) error {
	mtype := m.Type()

	if mtype.NumOut() != 0 {
		return errors.New("method should not have return values")
	}

	if mtype.NumIn() > 1 {
		return errors.New("method should not have maximum 1 arg")
	}

	if mtype.NumIn() == 0 {
		m.Call(nil)
		return nil
	}

	argt := mtype.In(0)
	argv := reflect.New(argt)
	argi := argv.Interface()

	if err := json.Unmarshal([]byte(jval), argi); err != nil {
		return errors.Wrap(err, "mapping method 1st arg failed")
	}

	arg := argv.Elem()
	m.Call([]reflect.Value{arg})
	return nil
}

func assignComponentField(f reflect.Value, jval string) error {
	fv := f.Addr()
	fi := fv.Interface()
	return json.Unmarshal([]byte(jval), fi)
}
