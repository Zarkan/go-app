package app

// UIStack is the interface that describes a container that displays its items
// as stacked panels.
type UIStack interface {
	UI

	// Center aligns the items from the center.
	Center() UIStack

	// Class adds a CSS class to the layout.
	Class(c string) UIStack

	// Content sets the content with the given UI elements.
	Content(elems ...UI) UIStack

	// End aligns the items from the end.
	End() UIStack

	// Stretch tries to make the items occupy all the space.
	Stretch() UIStack

	// Vertical stacks items vertically.
	Vertical() UIStack
}

// Stack creates a container that displays its items as stacked panels.
func Stack() UIStack {
	return &stack{
		Ialignment: "flex-start",
		Idirection: "row",
	}
}

type stack struct {
	Compo

	Ialignment string
	Iclass     string
	Idirection string
	Icontent   []UI
}

func (s *stack) Center() UIStack {
	s.Ialignment = "center"
	return s
}

func (s *stack) Class(c string) UIStack {
	if s.Iclass != "" {
		s.Iclass += " "
	}

	s.Iclass += c
	return s
}

func (s *stack) Content(elems ...UI) UIStack {
	s.Icontent = FilterUIElems(elems...)
	return s
}

func (s *stack) End() UIStack {
	s.Ialignment = "flex-end"
	return s
}

func (s *stack) Stretch() UIStack {
	s.Ialignment = "stretch"
	return s
}

func (s *stack) Vertical() UIStack {
	s.Idirection = "column"
	return s
}

func (s *stack) Render() UI {
	return Div().
		Class(s.Iclass).
		Style("position", "relative").
		Style("display", "flex").
		Style("flex-direction", s.Idirection).
		Style("align-items", s.Ialignment).
		Body(s.Icontent...)
}
