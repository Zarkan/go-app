package core

import (
	"path/filepath"

	"github.com/murlokswarm/app"
)

// Driver is a base struct to embed in app.Driver implementations.
type Driver struct {
}

// Run satisfies the app.Driver interface.
func (d *Driver) Run(f app.Factory) error {
	return app.ErrNotSupported
}

// AppName satisfies the app.Driver interface.
func (d *Driver) AppName() string {
	return ""
}

// Resources satisfies the app.Driver interface.
func (d *Driver) Resources(p ...string) string {
	r := filepath.Join(p...)
	r = filepath.Join("resources", r)
	return r
}

// Storage satisfies the app.Driver interface.
func (d *Driver) Storage(p ...string) string {
	s := filepath.Join(p...)
	s = filepath.Join("storage", s)
	return s
}

// Render satisfies the app.Driver interface.
func (d *Driver) Render(c app.Compo) error {
	return app.ErrNotSupported
}

// ElemByCompo satisfies the app.Driver interface.
func (d *Driver) ElemByCompo(c app.Compo) app.Elem {
	return &Elem{err: app.ErrNotSupported}
}

// NewWindow satisfies the app.Driver interface.
func (d *Driver) NewWindow(c app.WindowConfig) app.Window {
	w := &Window{}
	w.SetErr(app.ErrNotSupported)
	return w
}

// NewContextMenu satisfies the app.Driver interface.
func (d *Driver) NewContextMenu(c app.MenuConfig) app.Menu {
	m := &Menu{}
	m.SetErr(app.ErrNotSupported)
	return m
}

// NewPage satisfies the app.Driver interface.
func (d *Driver) NewPage(c app.PageConfig) error {
	return app.ErrNotSupported
}

// NewFilePanel satisfies the app.Driver interface.
func (d *Driver) NewFilePanel(c app.FilePanelConfig) error {
	return app.ErrNotSupported
}

// NewSaveFilePanel satisfies the app.Driver interface.
func (d *Driver) NewSaveFilePanel(c app.SaveFilePanelConfig) error {
	return app.ErrNotSupported
}

// NewShare satisfies the app.Driver interface.
func (d *Driver) NewShare(v interface{}) error {
	return app.ErrNotSupported
}

// NewNotification satisfies the app.Driver interface.
func (d *Driver) NewNotification(c app.NotificationConfig) app.Elem {
	return &Elem{err: app.ErrNotSupported}
}

// MenuBar satisfies the app.Driver interface.
func (d *Driver) MenuBar() app.Menu {
	m := &Menu{}
	m.SetErr(app.ErrNotSupported)
	return m
}

// NewStatusMenu satisfies the app.Driver interface.
func (d *Driver) NewStatusMenu(c app.StatusMenuConfig) app.StatusMenu {
	s := &StatusMenu{}
	s.SetErr(app.ErrNotSupported)
	return s
}

// Dock satisfies the app.Driver interface.
func (d *Driver) Dock() app.DockTile {
	dt := &DockTile{}
	dt.SetErr(app.ErrNotSupported)
	return dt
}

// CallOnUIGoroutine satisfies the app.Driver interface.
func (d *Driver) CallOnUIGoroutine(f func()) {
	f()
}
