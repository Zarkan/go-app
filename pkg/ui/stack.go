package ui

import "github.com/maxence-charriere/go-app/v9/pkg/app"

// IStack is the interface that describes a container that displays its items
// as stacked panels.
type IStack interface {
	app.UI

	// Sets the ID.
	ID(v string) IStack

	// Sets the class. Multiple classes can be defined by successive calls.
	Class(v string) IStack

	// Left aligns the content on the left.
	Left() IStack

	// Center aligns the content on the horizontal center.
	Center() IStack

	// Right aligns the content on the right.
	Right() IStack

	// Top aligns the content on the top.
	Top() IStack

	// Middle aligns the content on the vertical center.
	Middle() IStack

	// Bottom aligns the content on the bottom.
	Bottom() IStack

	// Stretch stretches the content vertically.
	Stretch() IStack

	// Content sets the content with the given UI elements.
	Content(elems ...app.UI) IStack
}

// Stack creates a container that displays its items as stacked panels.
func Stack() IStack {
	return &stack{
		IhorizontalAlign: "flex-start",
		IverticalAlign:   "flex-start",
	}
}

type stack struct {
	app.Compo

	Iid              string
	Iclass           string
	IhorizontalAlign string
	IverticalAlign   string
	Icontent         []app.UI
}

func (s *stack) ID(v string) IStack {
	s.Iid = v
	return s
}

func (s *stack) Left() IStack {
	s.IhorizontalAlign = "flex-start"
	return s
}

func (s *stack) Center() IStack {
	s.IhorizontalAlign = "center"
	return s
}

func (s *stack) Right() IStack {
	s.IhorizontalAlign = "flex-end"
	return s
}

func (s *stack) Top() IStack {
	s.IverticalAlign = "flex-start"
	return s
}

func (s *stack) Middle() IStack {
	s.IverticalAlign = "center"
	return s
}

func (s *stack) Bottom() IStack {
	s.IverticalAlign = "flex-end"
	return s
}

func (s *stack) Stretch() IStack {
	s.IverticalAlign = "stretch"
	return s
}

func (s *stack) Class(v string) IStack {
	if v == "" {
		return s
	}
	if s.Iclass != "" {
		s.Iclass += " "
	}
	s.Iclass += v
	return s
}

func (s *stack) Content(elems ...app.UI) IStack {
	s.Icontent = app.FilterUIElems(elems...)
	return s
}

func (s *stack) Render() app.UI {
	return app.Div().
		DataSet("goapp", "Stack").
		ID(s.Iid).
		Class(s.Iclass).
		Style("display", "flex").
		Style("justify-content", s.IhorizontalAlign).
		Style("align-items", s.IverticalAlign).
		Body(s.Icontent...)
}
