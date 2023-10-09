package app

import (
	"fmt"
	"html"
	"io"

	"github.com/maxence-charriere/go-app/v9/pkg/errors"
)

// Text creates a simple text element.
func Text(v any) UI {
	return &text{value: toString(v)}
}

// Text creates a simple text element with the given format and values.
func Textf(format string, v ...any) UI {
	return &text{value: fmt.Sprintf(format, v...)}
}

type text struct {
	disp          Dispatcher
	jsvalue       Value
	parentElement UI
	value         string
}

func (t *text) JSValue() Value {
	return t.jsvalue
}

func (t *text) Mounted() bool {
	return t.jsvalue != nil
}

func (t *text) name() string {
	return "text"
}

func (t *text) self() UI {
	return t
}

func (t *text) setSelf(n UI) {
}

func (t *text) getDispatcher() Dispatcher {
	return t.disp
}

func (t *text) getAttributes() attributes {
	return nil
}

func (t *text) getEventHandlers() eventHandlers {
	return nil
}

func (t *text) getParent() UI {
	return t.parentElement
}

func (t *text) setParent(p UI) UI {
	t.parentElement = p
	return t
}

func (t *text) getChildren() []UI {
	return nil
}

func (t *text) mount(d Dispatcher) error {
	if t.Mounted() {
		return errors.New("mounting ui element failed").
			WithTag("reason", "already mounted").
			WithTag("name", t.name()).
			WithTag("value", t.value)
	}

	t.disp = d
	t.jsvalue = Window().createTextNode(t.value)
	return nil
}

func (t *text) dismount() {
	t.jsvalue = nil
}

func (t *text) canUpdateWith(n UI) bool {
	_, ok := n.(*text)
	return ok
}

func (t *text) updateWith(n UI) error {
	if !t.Mounted() {
		return nil
	}

	o, _ := n.(*text)
	if t.value != o.value {
		t.value = o.value
		t.JSValue().setNodeValue(o.value)
	}

	return nil
}

func (t *text) onComponentEvent(any) {
}

func (t *text) html(w io.Writer) {
	w.Write([]byte(html.EscapeString(t.value)))
}

func (t *text) htmlWithIndent(w io.Writer, indent int) {
	writeIndent(w, indent)
	w.Write([]byte(html.EscapeString(t.value)))
}

func (t *text) parent() UI {
	return t.parentElement
}
