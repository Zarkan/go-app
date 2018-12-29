package app_test

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/test"
	"github.com/murlokswarm/app/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	app.Import(&tests.Foo{})

	defer func() { recover() }()
	app.Import(tests.NoPointerCompo{})
}

func TestApp(t *testing.T) {
	app.Logger = func(format string, a ...interface{}) {
		log := fmt.Sprintf(format, a...)
		t.Log(log)
	}

	app.Import(&tests.Foo{})
	app.Import(&tests.Bar{})

	app.Addons(app.Logs())

	onRun := func() {
		d := app.CurrentDriver()
		require.NotNil(t, d)

		assert.NotEmpty(t, app.Name())
		assert.Equal(t, filepath.Join("resources", "hello", "world"), app.Resources("hello", "world"))
		assert.Equal(t, filepath.Join("storage", "hello", "world"), app.Storage("hello", "world"))

		app.Render(&tests.Hello{})
		assert.NotNil(t, app.ElemByCompo(&tests.Hello{}))

		assert.NotNil(t, app.NewWindow(app.WindowConfig{}))
		assert.NotNil(t, app.NewPage(app.PageConfig{}))
		assert.NotNil(t, app.NewContextMenu(""))
		assert.NotNil(t, app.NewController(app.ControllerConfig{}))
		assert.NotNil(t, app.NewFilePanel(app.FilePanelConfig{}))
		assert.NotNil(t, app.NewSaveFilePanel(app.SaveFilePanelConfig{}))
		assert.NotNil(t, app.NewShare("boo"))
		assert.NotNil(t, app.NewNotification(app.NotificationConfig{}))
		assert.NotNil(t, app.MenuBar())
		assert.NotNil(t, app.NewStatusMenu(app.StatusMenuConfig{}))
		assert.NotNil(t, app.Dock())
		assert.NotNil(t, app.NewStatusMenu(app.StatusMenuConfig{}))

		app.Emit("test")

		app.UI(func() {
			app.Logf("hello")
		})

		go time.AfterFunc(time.Millisecond, app.Stop)
	}

	defer app.NewSubscriber().
		Subscribe(app.Running, onRun).
		Close()

	err := app.Run()
	assert.Error(t, err)
	err = app.Run(&test.Driver{})
	assert.Error(t, err)
}

func TestLog(t *testing.T) {
	log := ""

	app.Logger = func(format string, a ...interface{}) {
		log = fmt.Sprintf(format, a...)
	}

	app.Log("hello", "world")
	assert.Equal(t, "hello world", log)

	app.Logf("%s %s", "bye", "world")
	assert.Equal(t, "bye world", log)
}

func TestPanic(t *testing.T) {
	log := ""

	app.Logger = func(format string, a ...interface{}) {
		log = fmt.Sprintf(format, a...)
	}

	defer func() {
		err := recover()
		assert.Equal(t, "hello world", log)
		assert.Equal(t, "hello world", err)
	}()

	app.Panic("hello", "world")
	assert.Fail(t, "no panic")
}

func TestPanicf(t *testing.T) {
	log := ""

	app.Logger = func(format string, a ...interface{}) {
		log = fmt.Sprintf(format, a...)
	}

	defer func() {
		err := recover()
		assert.Equal(t, "bye world", log)
		assert.Equal(t, "bye world", err)
	}()

	app.Panicf("%s %s", "bye", "world")
	assert.Fail(t, "no panic")
}

func TestPretty(t *testing.T) {
	t.Log(app.Pretty(app.WindowConfig{}))
}
