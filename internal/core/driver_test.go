package core

import (
	"testing"

	"github.com/murlokswarm/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriver(t *testing.T) {
	app.Logger = t.Logf

	d := &Driver{}
	require.Implements(t, (*app.Driver)(nil), d)

	assert.Empty(t, d.Target())
	assert.Error(t, d.Run(app.DriverConfig{}))
	assert.NotEmpty(t, d.AppName())
	assert.Equal(t, "resources", d.Resources())
	assert.Equal(t, "storage", d.Storage())
	d.Render(nil)
	assert.Error(t, d.ElemByCompo(nil).Err())

	w := d.NewWindow(app.WindowConfig{})
	assert.Error(t, w.Err())

	m := d.NewContextMenu(app.MenuConfig{})
	assert.Error(t, m.Err())

	fp := d.NewFilePanel(app.FilePanelConfig{})
	assert.Error(t, fp.Err())

	fsp := d.NewSaveFilePanel(app.SaveFilePanelConfig{})
	assert.Error(t, fsp.Err())

	s := d.NewShare(nil)
	assert.Error(t, s.Err())

	n := d.NewNotification(app.NotificationConfig{})
	assert.Error(t, n.Err())

	mb := d.MenuBar()
	assert.Error(t, mb.Err())

	c := d.NewController(app.ControllerConfig{})
	assert.Error(t, c.Err())

	sm := d.NewStatusMenu(app.StatusMenuConfig{})
	assert.Error(t, sm.Err())

	dt := d.DockTile()
	assert.Error(t, dt.Err())

	d.UI(func() {
		t.Log("call from ui goroutine")
	})

	d.Stop()
}
