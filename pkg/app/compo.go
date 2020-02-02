package app

import (
	"fmt"
	"reflect"
	"unicode"
)

// Compo represents the base struct to use in order to build a component.
type Compo struct {
	compo      CompoNode
	parentNode nodeWithChildren
	root       ValueNode
}

func (c *Compo) nodeType() reflect.Type {
	return reflect.TypeOf(c.compo)
}

// JSValue returns the component root javascript value.
func (c *Compo) JSValue() Value {
	return c.root.JSValue()
}

func (c *Compo) parent() nodeWithChildren {
	return c.parentNode
}

func (c *Compo) setParent(p nodeWithChildren) {
	c.parentNode = p
}

func (c *Compo) setCompo(n CompoNode) {
	c.compo = n
}

func (c *Compo) dismount() {
	c.root.dismount()

	if dismounter, ok := c.root.(Dismounter); ok {
		dismounter.OnDismount()
	}
}

func (c *Compo) replaceChild(old, new ValueNode) {
	if old == c.root {
		c.root = new
	}
}

// Update update the component appearance. It should be called when a field
// used to render the component has been modified.
func (c *Compo) Update() {
	current := c.root
	incoming := c.compo.Render().(ValueNode)

	if err := update(current, incoming); err != nil {
		panic(fmt.Errorf("%T: updating component failed: %w", c.compo, err))
	}
}

func (c *Compo) mount(compo CompoNode) error {
	c.setCompo(compo)

	root := compo.Render()
	if err := mount(root); err != nil {
		return fmt.Errorf("%T: invalid root: %w", compo, err)
	}
	c.root = root

	if mounter, ok := compo.(Mounter); ok {
		mounter.OnMount()
	}

	return nil
}

func (c *Compo) update(n CompoNode) {
	aval := reflect.Indirect(reflect.ValueOf(c.compo))
	bval := reflect.Indirect(reflect.ValueOf(n))
	compotype := reflect.ValueOf(c).Elem().Type()

	for i := 0; i < aval.NumField(); i++ {
		a := aval.Field(i)
		b := bval.Field(i)

		if a.Type() == compotype {
			continue
		}

		if !isExported(aval.Type().Field(i).Name) {
			continue
		}

		if !reflect.DeepEqual(a.Interface(), b.Interface()) && a.CanSet() {
			a.Set(b)
		}
	}
}

func isExported(fieldOrMethod string) bool {
	return !unicode.IsLower(rune(fieldOrMethod[0]))
}