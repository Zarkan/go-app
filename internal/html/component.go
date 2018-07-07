package html

import (
	"bytes"
	"encoding/json"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/murlokswarm/app"
	"github.com/pkg/errors"
)

type compoRow struct {
	id        string
	component app.Component
	events    app.EventSubscriber
	root      node
}

func validateComponent(c app.Component) error {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr {
		return errors.New("component is not a pointer")
	}

	v = v.Elem()
	if v.NumField() == 0 {
		return errors.New("component is based on a struct without field. use app.ZeroCompo instead of struct{}")
	}
	return nil
}

func decodeComponent(c app.Component) (node, error) {
	var funcs template.FuncMap

	if compoExtRend, ok := c.(app.ComponentWithExtendedRender); ok {
		funcs = compoExtRend.Funcs()
	}

	if len(funcs) == 0 {
		funcs = make(template.FuncMap, 4)
	}

	funcs["raw"] = func(s string) template.HTML {
		return template.HTML(s)
	}

	funcs["compo"] = func(s string) template.HTML {
		return template.HTML("<" + s + ">")
	}

	funcs["time"] = func(t time.Time, layout string) string {
		return t.Format(layout)
	}

	funcs["json"] = func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	}

	tmpl, err := template.
		New("").
		Funcs(funcs).
		Parse(c.Render())
	if err != nil {
		return nil, err
	}

	var w bytes.Buffer
	if err = tmpl.Execute(&w, c); err != nil {
		return nil, err
	}
	return decodeNodes(w.String())
}

func mapComponentFields(c app.Component, fields map[string]string) error {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()

	for i, numfields := 0, t.NumField(); i < numfields; i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if ft.Anonymous {
			continue
		}

		// Ignore non exported field.
		if len(ft.PkgPath) != 0 {
			continue
		}

		name := strings.ToLower(ft.Name)
		value, ok := fields[name]

		// Remove not set boolean.
		if !ok && fv.Kind() == reflect.Bool {
			fv.SetBool(false)
			continue
		} else if !ok {
			continue
		}

		if err := mapComponentField(fv, value); err != nil {
			return err
		}
	}
	return nil
}

func mapComponentField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		if len(value) == 0 {
			value = "true"
		}
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		n, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return err
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uintptr:
		n, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return err
		}
		field.SetUint(n)

	case reflect.Float64, reflect.Float32:
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(n)

	default:
		addr := field.Addr()
		i := addr.Interface()
		if err := json.Unmarshal([]byte(value), i); err != nil {
			return err
		}
	}
	return nil
}
