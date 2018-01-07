// +build darwin,amd64

package mac

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/appjs"
	"github.com/murlokswarm/app/bridge"
	"github.com/murlokswarm/app/html"
	"github.com/pkg/errors"
)

// Window implements the app.Window interface.
type Window struct {
	id        uuid.UUID
	markup    app.Markup
	component app.Component
	history   app.History
	lastFocus time.Time

	onMove           func(x, y float64)
	onResize         func(width, height float64)
	onFocus          func()
	onBlur           func()
	onFullScreen     func()
	onExitFullScreen func()
	onMinimize       func()
	onDeminimize     func()
	onClose          func() bool
}

func newWindow(config app.WindowConfig) (w *Window, err error) {
	var markup app.Markup = html.NewMarkup(driver.factory)
	markup = app.NewConcurrentMarkup(markup)

	history := app.NewHistory()
	history = app.NewConcurrentHistory(history)

	w = &Window{
		id:        uuid.New(),
		markup:    markup,
		history:   history,
		lastFocus: time.Now(),

		onMove:           config.OnMove,
		onResize:         config.OnResize,
		onFocus:          config.OnFocus,
		onBlur:           config.OnBlur,
		onFullScreen:     config.OnFullScreen,
		onExitFullScreen: config.OnExitFullScreen,
		onMinimize:       config.OnMinimize,
		onDeminimize:     config.OnDeminimize,
		onClose:          config.OnClose,
	}

	config.MinWidth, config.MaxWidth = normalizeWidowSize(config.MinWidth, config.MaxWidth)
	config.MinHeight, config.MaxHeight = normalizeWidowSize(config.MinHeight, config.MaxHeight)

	if _, err = driver.macos.Request(
		fmt.Sprintf("/window/new?id=%s", w.id),
		bridge.NewPayload(config),
	); err != nil {
		return
	}

	if err = driver.elements.Add(w); err != nil {
		return
	}

	if len(config.DefaultURL) != 0 {
		err = w.Load(config.DefaultURL)
	}
	return
}

func normalizeWidowSize(min, max float64) (float64, float64) {
	min = math.Max(0, min)
	min = math.Min(min, 10000)

	if max == 0 {
		max = 10000
	}
	max = math.Max(0, max)
	max = math.Min(max, 10000)

	min = math.Min(min, max)
	return min, max
}

// ID satisfies the app.Element interface.
func (w *Window) ID() uuid.UUID {
	return w.id
}

// Load satisfies the app.ElementWithComponent interface.
func (w *Window) Load(rawurl string, v ...interface{}) error {
	rawurl = fmt.Sprintf(rawurl, v...)
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	compoName := app.ComponentNameFromURL(u)
	isRegiteredCompo := driver.factory.IsRegisteredComponent(compoName)
	currentURL, err := w.history.Current()

	if isRegiteredCompo && (err != nil || currentURL != u.String()) {
		w.history.NewEntry(u.String())
	}
	return w.load(u)
}

func (w *Window) load(u *url.URL) error {
	var compoName = app.ComponentNameFromURL(u)
	var compo app.Component
	var root app.Tag
	var buffer bytes.Buffer
	var pageConfig html.PageConfig
	var err error

	// Redirect web page to default web browser.
	if compo, err = driver.factory.NewComponent(compoName); err != nil {
		cmd := exec.Command("open", u.String())
		err = cmd.Run()
		return err
	}

	if w.component != nil {
		w.markup.Dismount(w.component)
	}

	if _, err = w.markup.Mount(compo); err != nil {
		return err
	}
	w.component = compo

	if navigable, ok := compo.(app.Navigable); ok {
		navigable.OnNavigate(u)
	}

	if root, err = w.markup.Root(compo); err != nil {
		return nil
	}

	enc := html.NewEncoder(&buffer, w.markup)
	if err = enc.Encode(root); err != nil {
		return err
	}

	if page, ok := compo.(html.Page); ok {
		pageConfig = page.PageConfig()
	}
	if len(pageConfig.CSS) == 0 {
		pageConfig.CSS = defaultCSS()
	}
	pageConfig.DefaultComponent = template.HTML(buffer.String())
	pageConfig.AppJS = appjs.AppJS("window.webkit.messageHandlers.golangRequest.postMessage")

	payload := struct {
		Title   string `json:"title"`
		Page    string `json:"page"`
		LoadURL string `json:"load-url"`
		BaseURL string `json:"base-url"`
	}{
		Title:   pageConfig.Title,
		Page:    html.NewPage(pageConfig),
		LoadURL: u.String(),
		BaseURL: driver.Resources(),
	}

	_, err = driver.macos.RequestWithAsyncResponse(
		fmt.Sprintf("/window/load?id=%s", w.id),
		bridge.NewPayload(payload),
	)
	return err
}

func defaultCSS() (css []string) {
	cssdir := filepath.Join(app.Resources(), "css")

	filepath.Walk(cssdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext := filepath.Ext(path); ext != ".css" {
			return nil
		}

		css = append(css, path)
		return nil
	})
	return
}

// Contains satisfies the app.ElementWithComponent interface.
func (w *Window) Contains(compo app.Component) bool {
	return w.markup.Contains(compo)
}

// Render satisfies the app.ElementWithComponent interface.
func (w *Window) Render(compo app.Component) error {
	syncs, err := w.markup.Update(compo)
	if err != nil {
		return err
	}

	for _, sync := range syncs {
		if sync.Replace {
			err = w.render(sync)
		} else {
			err = w.renderAttributes(sync)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Window) render(sync app.TagSync) error {
	var buffer bytes.Buffer

	enc := html.NewEncoder(&buffer, w.markup)
	if err := enc.Encode(sync.Tag); err != nil {
		return err
	}

	payload := struct {
		ID        string `json:"id"`
		Component string `json:"component"`
	}{
		ID:        sync.Tag.ID.String(),
		Component: buffer.String(),
	}

	_, err := driver.macos.Request(
		fmt.Sprintf("/window/render?id=%s", w.id),
		bridge.NewPayload(payload),
	)
	return err
}

func (w *Window) renderAttributes(sync app.TagSync) error {
	payload := struct {
		ID         string           `json:"id"`
		Attributes app.AttributeMap `json:"attributes"`
	}{
		ID:         sync.Tag.ID.String(),
		Attributes: sync.Tag.Attributes,
	}

	_, err := driver.macos.Request(
		fmt.Sprintf("/window/render/attributes?id=%s", w.id),
		bridge.NewPayload(payload),
	)
	return err
}

// LastFocus satisfies the app.ElementWithComponent interface.
func (w *Window) LastFocus() time.Time {
	return w.lastFocus
}

// Reload satisfies the app.ElementWithNavigation interface.
func (w *Window) Reload() error {
	var rawurl string
	var u *url.URL
	var err error

	if rawurl, err = w.history.Current(); err != nil {
		return err
	}

	if u, err = url.Parse(rawurl); err != nil {
		return err
	}

	return w.load(u)
}

// CanPrevious satisfies the app.ElementWithNavigation interface.
func (w *Window) CanPrevious() bool {
	return w.history.CanPrevious()
}

// Previous satisfies the app.ElementWithNavigation interface.
func (w *Window) Previous() error {
	var rawurl string
	var u *url.URL
	var err error

	if rawurl, err = w.history.Previous(); err != nil {
		return err
	}

	if u, err = url.Parse(rawurl); err != nil {
		return err
	}

	return w.load(u)
}

// CanNext satisfies the app.ElementWithNavigation interface.
func (w *Window) CanNext() bool {
	return w.history.CanNext()
}

// Next satisfies the app.ElementWithNavigation interface.
func (w *Window) Next() error {
	var rawurl string
	var u *url.URL
	var err error

	if rawurl, err = w.history.Next(); err != nil {
		return err
	}

	if u, err = url.Parse(rawurl); err != nil {
		return err
	}

	return w.load(u)
}

// Position satisfies the app.Window interface.
func (w *Window) Position() (x, y float64) {
	rawurl := fmt.Sprintf("/window/position?id=%s", w.id)

	res, err := driver.macos.RequestWithAsyncResponse(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "retrieving positon of window %v", w.ID()))
	}

	var pos point
	res.Unmarshal(&pos)
	return pos.X, pos.Y
}

// Move satisfies the app.Window interface.
func (w *Window) Move(x, y float64) {
	rawurl := fmt.Sprintf("/window/move?id=%s", w.id)
	payload := bridge.NewPayload(point{
		X: x,
		Y: y,
	})

	_, err := driver.macos.Request(rawurl, payload)
	if err != nil {
		panic(errors.Wrapf(err, "moving window %v failed", w.ID()))
	}
}

func onWindowMove(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onMove == nil {
		return
	}

	var pos point
	p.Unmarshal(&pos)
	w.onMove(pos.X, pos.Y)
	return
}

// Center satisfies the app.Window interface.
func (w *Window) Center() {
	rawurl := fmt.Sprintf("/window/center?id=%s", w.id)

	_, err := driver.macos.Request(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "centering window %v failed", w.ID()))
	}
}

// Size satisfies the app.Window interface.
func (w *Window) Size() (width, height float64) {
	rawurl := fmt.Sprintf("/window/size?id=%s", w.id)

	res, err := driver.macos.RequestWithAsyncResponse(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "retrieving size of window %v failed", w.ID()))
	}

	var size size
	res.Unmarshal(&size)
	return size.Width, size.Height
}

// Resize satisfies the app.Window interface.
func (w *Window) Resize(width, height float64) {
	rawurl := fmt.Sprintf("/window/resize?id=%s", w.id)
	payload := bridge.NewPayload(size{
		Width:  width,
		Height: height,
	})

	_, err := driver.macos.Request(rawurl, payload)
	if err != nil {
		panic(errors.Wrapf(err, "resizing window %v failed", w.ID()))
	}
}

func onWindowResize(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onResize == nil {
		return
	}

	var size size
	p.Unmarshal(&size)
	w.onResize(size.Width, size.Height)
	return
}

// Focus satisfies the app.Window interface.
func (w *Window) Focus() {
	rawurl := fmt.Sprintf("/window/focus?id=%s", w.id)

	_, err := driver.macos.Request(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "focusing window %v failed", w.ID()))
	}
}

func onWindowFocus(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	w.lastFocus = time.Now()

	if w.onFocus == nil {
		return
	}

	w.onFocus()
	return
}

func onWindowBlur(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onBlur == nil {
		return
	}

	w.onBlur()
	return
}

// ToggleFullScreen satisfies the app.Window interface.
func (w *Window) ToggleFullScreen() {
	rawurl := fmt.Sprintf("/window/togglefullscreen?id=%s", w.id)

	_, err := driver.macos.Request(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "toggling full screen on window %v failed", w.ID()))
	}
}

func onWindowFullScreen(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onFullScreen == nil {
		return
	}

	w.onFullScreen()
	return
}

func onWindowExitFullScreen(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onExitFullScreen == nil {
		return
	}

	w.onExitFullScreen()
	return
}

// ToggleMinimize satisfies the app.Window interface.
func (w *Window) ToggleMinimize() {
	rawurl := fmt.Sprintf("/window/toggleminimize?id=%s", w.id)

	_, err := driver.macos.Request(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "toggling minimize on window %v failed", w.ID()))
	}
}

func onWindowMinimize(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onMinimize == nil {
		return
	}

	w.onMinimize()
	return
}

func onWindowDeminimize(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	if w.onDeminimize == nil {
		return
	}

	w.onDeminimize()
	return
}

// Close satisfies the app.Window interface.
func (w *Window) Close() {
	rawurl := fmt.Sprintf("/window/close?id=%s", w.id)

	_, err := driver.macos.Request(rawurl, nil)
	if err != nil {
		panic(errors.Wrapf(err, "closing window %v failed", w.ID()))
	}
}

func onWindowClose(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	shouldClose := true
	if w.onClose != nil {
		shouldClose = w.onClose()
	}
	res = bridge.NewPayload(shouldClose)

	if w.component != nil {
		w.markup.Dismount(w.component)
	}

	if shouldClose {
		driver.elements.Remove(w)
	}
	return
}

func onWindowCallback(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	var mapping app.Mapping
	p.Unmarshal(&mapping)

	function, err := w.markup.Map(mapping)
	if err != nil {
		app.DefaultLogger.Error(err)
		return
	}

	if function != nil {
		function()
		return
	}

	var compo app.Component
	if compo, err = w.markup.Component(mapping.CompoID); err != nil {
		app.DefaultLogger.Error(err)
		return
	}

	if err = w.Render(compo); err != nil {
		app.DefaultLogger.Error(err)
	}
	return
}

func onWindowNavigate(w *Window, u *url.URL, p bridge.Payload) (res bridge.Payload) {
	var rawurl string
	p.Unmarshal(&rawurl)
	w.Load(rawurl)
	return
}
