package app

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maxence-charriere/go-app/v9/pkg/errors"
)

// Context represents a context that is tied to a UI element. It is canceled
// when the element is dismounted.
//
// It implements the context.Context interface.
//  https://golang.org/pkg/context/#Context
type Context struct {
	context.Context

	// The UI element tied to the context.
	Src UI

	// The JavaScript value of the element tied to the context. This is a
	// shorthand for:
	//  ctx.Src.JSValue()
	JSSrc Value

	// Reports whether the app has been updated in background. Use app.Reload()
	// to load the updated version.
	AppUpdateAvailable bool

	// The info about the current page.
	Page Page

	dispatcher Dispatcher
}

// Dispatch executes the given function on the UI goroutine and notifies the
// context's nearest component to update its state.
func (ctx Context) Dispatch(fn func(Context)) {
	ctx.dispatcher.Dispatch(ctx.Src, fn)
}

// Defer executes the given function on the UI goroutine after notifying the
// context's nearest component to update its state.
func (ctx Context) Defer(fn func(Context)) {
	ctx.dispatcher.Defer(ctx.Src, fn)
}

// Async executes the given function on a new goroutine.
//
// The difference versus just launching a goroutine is that it ensures that the
// asynchronous function is called before a page is fully pre-rendered and
// served over HTTP.
func (ctx Context) Async(fn func()) {
	ctx.dispatcher.Async(fn)
}

// Emit executes the given function and notifies the parent components to update
// their state. It should be used to launch component custom event handlers.
func (ctx Context) Emit(fn func()) {
	ctx.dispatcher.Emit(ctx.Src, fn)
}

// Reload reloads the WebAssembly app at the current page.
func (ctx Context) Reload() {
	if IsServer {
		return
	}

	ctx.Defer(func(ctx Context) {
		Window().Get("location").Call("reload")
	})
}

// Navigate navigates to the given URL. This is a helper method that converts
// rawURL to an *url.URL and then calls ctx.NavigateTo under the hood.
func (ctx Context) Navigate(rawURL string) {
	ctx.Defer(func(ctx Context) {
		navigate(ctx.dispatcher, rawURL)
	})
}

// NavigateTo navigates to the given URL.
func (ctx Context) NavigateTo(u *url.URL) {
	ctx.Defer(func(ctx Context) {
		navigateTo(ctx.dispatcher, u, true)
	})
}

// ResolveStaticResource resolves the given path to make it point to the right
// location whether static resources are located on a local directory or a
// remote bucket.
func (ctx Context) ResolveStaticResource(path string) string {
	return ctx.dispatcher.resolveStaticResource(path)
}

// LocalStorage returns a storage that uses the browser local storage associated
// to the document origin. Data stored has no expiration time.
func (ctx Context) LocalStorage() BrowserStorage {
	return ctx.dispatcher.localStorage()
}

// SessionStorage returns a storage that uses the browser session storage
// associated to the document origin. Data stored expire when the page
// session ends.
func (ctx Context) SessionStorage() BrowserStorage {
	return ctx.dispatcher.sessionStorage()
}

// ScrollTo scrolls to the HTML element with the given id.
func (ctx Context) ScrollTo(id string) {
	ctx.Defer(func(ctx Context) {
		Window().ScrollToID(id)
	})
}

// After asynchronously waits for the given duration and dispatches the given
// function.
func (ctx Context) After(d time.Duration, fn func(Context)) {
	ctx.Async(func() {
		time.Sleep(d)
		ctx.Dispatch(fn)
	})
}

// DeviceID returns a UUID that identifies the app on the current device.
func (ctx Context) DeviceID() string {
	var id string
	if err := ctx.LocalStorage().Get("/go-app/deviceID", &id); err != nil {
		panic(errors.New("retrieving device id failed").Wrap(err))
	}
	if id != "" {
		return id
	}

	id = uuid.New().String()
	if err := ctx.LocalStorage().Set("/go-app/deviceID", id); err != nil {
		panic(errors.New("creating device id failed").Wrap(err))
	}
	return id
}

// Encrypt encrypts the given value using AES encryption.
func (ctx Context) Encrypt(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("encoding value failed").Wrap(err)
	}

	b, err = encrypt(ctx.cryptoKey(), b)
	if err != nil {
		return nil, errors.New("encrypting value failed").Wrap(err)
	}
	return b, nil
}

// Decrypt decrypts the given encrypted bytes and stores them in the given
// value.
func (ctx Context) Decrypt(crypted []byte, v interface{}) error {
	b, err := decrypt(ctx.cryptoKey(), crypted)
	if err != nil {
		return errors.New("decrypting value failed").Wrap(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return errors.New("decoding value failed").Wrap(err)
	}
	return nil
}

func (ctx Context) cryptoKey() string {
	return strings.ReplaceAll(ctx.DeviceID(), "-", "")
}

func makeContext(src UI) Context {
	return Context{
		Context:            src.context(),
		Src:                src,
		JSSrc:              src.JSValue(),
		AppUpdateAvailable: appUpdateAvailable,
		Page:               src.dispatcher().currentPage(),
		dispatcher:         src.dispatcher(),
	}
}
