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

// Context is the interface that describes a context tied to a UI element.
//
// A context provides mechanisms to deal with the browser, the current page,
// navigation, concurrency, and component communication.
//
// It is canceled when its associated UI element is dismounted.
type Context interface {
	context.Context

	// Returns the UI element tied to the context.
	Src() UI

	// Returns the associated JavaScript value. The is an helper method for:
	//  ctx.Src.JSValue()
	JSSrc() Value

	// Reports whether the app has been updated in background. Use app.Reload()
	// to load the updated version.
	AppUpdateAvailable() bool

	// Reports whether the app is installable.
	IsAppInstallable() bool

	// Shows the app install prompt if the app is installable.
	ShowAppInstallPrompt()

	// Returns the current page.
	Page() Page

	// Executes the given function on the UI goroutine and notifies the
	// context's nearest component to update its state.
	Dispatch(fn func(Context))

	// Executes the given function on the UI goroutine after notifying the
	// context's nearest component to update its state.
	Defer(fn func(Context))

	// Registers the handler for the given action name. When an action occurs,
	// the handler is executed on the UI goroutine.
	Handle(actionName string, h ActionHandler)

	// Creates an action with optional tags, to be handled with Context.Handle.
	// Eg:
	//  ctx.NewAction("myAction")
	//  ctx.NewAction("myAction", app.T("purpose", "test"))
	//  ctx.NewAction("myAction", app.Tags{
	//      "foo": "bar",
	//      "hello": "world",
	//  })
	NewAction(name string, tags ...Tagger)

	// Creates an action with a value and optional tags, to be handled with
	// Context.Handle. Eg:
	//  ctx.NewActionWithValue("processValue", 42)
	//  ctx.NewActionWithValue("processValue", 42, app.T("type", "number"))
	//  ctx.NewActionWithValue("myAction", 42, app.Tags{
	//      "foo": "bar",
	//      "hello": "world",
	//  })
	NewActionWithValue(name string, v interface{}, tags ...Tagger)

	// Executes the given function on a new goroutine.
	//
	// The difference versus just launching a goroutine is that it ensures that
	// the asynchronous function is called before a page is fully pre-rendered
	// and served over HTTP.
	Async(fn func())

	// Asynchronously waits for the given duration and dispatches the given
	// function.
	After(d time.Duration, fn func(Context))

	// Executes the given function and notifies the parent components to update
	// their state. It should be used to launch component custom event handlers.
	Emit(fn func())

	// Reloads the WebAssembly app to the current page. It is like refreshing
	// the browser page.
	Reload()

	// Navigates to the given URL. This is a helper method that converts url to
	// an *url.URL and then calls ctx.NavigateTo under the hood.
	Navigate(url string)

	// Navigates to the given URL.
	NavigateTo(u *url.URL)

	// Resolves the given path to make it point to the right location whether
	// static resources are located on a local directory or a remote bucket.
	ResolveStaticResource(string) string

	// Returns a storage that uses the browser local storage associated to the
	// document origin. Data stored has no expiration time.
	LocalStorage() BrowserStorage

	// Returns a storage that uses the browser session storage associated to the
	// document origin. Data stored expire when the page session ends.
	SessionStorage() BrowserStorage

	// Scrolls to the HTML element with the given id.
	ScrollTo(id string)

	// Returns a UUID that identifies the app on the current device.
	DeviceID() string

	// Encrypts the given value using AES encryption.
	Encrypt(v interface{}) ([]byte, error)

	// Decrypts the given encrypted bytes and stores them in the given value.
	Decrypt(crypted []byte, v interface{}) error

	// Sets the state with the given value.
	// Example:
	//  ctx.SetState("/globalNumber", 42, Persistent)
	//
	// Options can be added to persist a state into the local storage, encrypt,
	// expire, or broadcast the state across browser tabs and windows.
	// Example:
	//  ctx.SetState("/globalNumber", 42, Persistent, Broadcast)
	SetState(state string, v interface{}, opts ...StateOption)

	// Stores the specified state value into the given receiver. Panics when the
	// receiver is not a pointer or nil.
	GetState(state string, recv interface{})

	// Deletes the given state. All value observations are stopped.
	DelState(state string)

	// Creates an observer that observes changes for the given state.
	// Example:
	//  type myComponent struct {
	//      app.Compo
	//
	//      number int
	//  }
	//
	//  func (c *myComponent) OnMount(ctx app.Context) {
	//      ctx.ObserveState("/globalNumber").Value(&c.number)
	//  }
	ObserveState(state string) Observer

	// Returns the app dispatcher.
	Dispatcher() Dispatcher

	// Requests the user whether the app can use notifications.
	RequestNotificationPermission() bool

	// Creates a user notification.
	NewNotification(n Notification)
}

type uiContext struct {
	context.Context

	src                UI
	jsSrc              Value
	appUpdateAvailable bool
	page               Page
	disp               Dispatcher
}

func (ctx uiContext) Src() UI {
	return ctx.src
}

func (ctx uiContext) JSSrc() Value {
	return ctx.jsSrc
}

func (ctx uiContext) AppUpdateAvailable() bool {
	return ctx.appUpdateAvailable
}

func (ctx uiContext) IsAppInstallable() bool {
	if Window().Get("goappIsAppInstallable").Truthy() {
		return Window().Call("goappIsAppInstallable").Bool()
	}
	return false
}

func (ctx uiContext) IsAppInstalled() bool {
	if Window().Get("goappIsAppInstalled").Truthy() {
		return Window().Call("goappIsAppInstalled").Bool()
	}
	return false
}

func (ctx uiContext) ShowAppInstallPrompt() {
	if ctx.IsAppInstallable() {
		Window().Call("goappShowInstallPrompt")
	}
}

func (ctx uiContext) Page() Page {
	return ctx.page
}

func (ctx uiContext) Dispatch(fn func(Context)) {
	ctx.Dispatcher().Dispatch(Dispatch{
		Mode:     Update,
		Source:   ctx.Src(),
		Function: fn,
	})
}

func (ctx uiContext) Defer(fn func(Context)) {
	ctx.Dispatcher().Dispatch(Dispatch{
		Mode:     Defer,
		Source:   ctx.Src(),
		Function: fn,
	})
}

func (ctx uiContext) Handle(actionName string, h ActionHandler) {
	ctx.Dispatcher().Handle(actionName, ctx.Src(), h)
}

func (ctx uiContext) NewAction(name string, tags ...Tagger) {
	ctx.NewActionWithValue(name, nil, tags...)
}

func (ctx uiContext) NewActionWithValue(name string, v interface{}, tags ...Tagger) {
	var tagMap Tags
	for _, t := range tags {
		if tagMap == nil {
			tagMap = t.Tags()
			continue
		}
		for k, v := range t.Tags() {
			tagMap[k] = v
		}
	}

	ctx.Dispatcher().Post(Action{
		Name:  name,
		Value: v,
		Tags:  tagMap,
	})
}

func (ctx uiContext) Async(fn func()) {
	ctx.Dispatcher().Async(fn)
}

func (ctx uiContext) After(d time.Duration, fn func(Context)) {
	ctx.Async(func() {
		time.Sleep(d)
		ctx.Dispatch(fn)
	})
}

func (ctx uiContext) Emit(fn func()) {
	ctx.Dispatcher().Emit(ctx.Src(), fn)
}

func (ctx uiContext) Reload() {
	if IsServer {
		return
	}
	ctx.Defer(func(ctx Context) {
		Window().Get("location").Call("reload")
	})
}

func (ctx uiContext) Navigate(rawURL string) {
	ctx.Defer(func(ctx Context) {
		navigate(ctx.Dispatcher(), rawURL)
	})
}

func (ctx uiContext) NavigateTo(u *url.URL) {
	ctx.Defer(func(ctx Context) {
		navigateTo(ctx.Dispatcher(), u, true)
	})
}

func (ctx uiContext) ResolveStaticResource(path string) string {
	return ctx.Dispatcher().resolveStaticResource(path)
}

func (ctx uiContext) LocalStorage() BrowserStorage {
	return ctx.Dispatcher().localStorage()
}

func (ctx uiContext) SessionStorage() BrowserStorage {
	return ctx.Dispatcher().sessionStorage()
}

func (ctx uiContext) ScrollTo(id string) {
	ctx.Defer(func(ctx Context) {
		Window().ScrollToID(id)
	})
}

func (ctx uiContext) DeviceID() string {
	var id string
	if err := ctx.LocalStorage().Get("/go-app/deviceID", &id); err != nil {
		panic(errors.New("retrieving device id failed").Wrap(err))
	}
	if id != "" {
		return id
	}

	id = uuid.NewString()
	if err := ctx.LocalStorage().Set("/go-app/deviceID", id); err != nil {
		panic(errors.New("creating device id failed").Wrap(err))
	}
	return id
}

func (ctx uiContext) Encrypt(v interface{}) ([]byte, error) {
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

func (ctx uiContext) Decrypt(crypted []byte, v interface{}) error {
	b, err := decrypt(ctx.cryptoKey(), crypted)
	if err != nil {
		return errors.New("decrypting value failed").Wrap(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return errors.New("decoding value failed").Wrap(err)
	}
	return nil
}

func (ctx uiContext) SetState(state string, v interface{}, opts ...StateOption) {
	ctx.Dispatcher().SetState(state, v, opts...)
}

func (ctx uiContext) GetState(state string, recv interface{}) {
	ctx.Dispatcher().GetState(state, recv)
}

func (ctx uiContext) DelState(state string) {
	ctx.Dispatcher().DelState(state)
}

func (ctx uiContext) ObserveState(state string) Observer {
	return ctx.Dispatcher().ObserveState(state, ctx.src)
}

func (ctx uiContext) Dispatcher() Dispatcher {
	return ctx.disp
}

func (ctx uiContext) RequestNotificationPermission() bool {
	notification := Window().Get("Notification")
	if !notification.Truthy() {
		return false
	}

	permission := make(chan string, 1)
	defer close(permission)

	notification.Call("requestPermission").Then(func(v Value) {
		permission <- v.String()
	})

	return <-permission == "granted"
}

func (ctx uiContext) NewNotification(n Notification) {
	setObjectField := func(obj map[string]interface{}, name string, value interface{}) {
		switch v := value.(type) {
		case string:
			if v == "" {
				return
			}
			switch name {
			case "badge", "icon", "image":
				obj[name] = ctx.Dispatcher().resolveStaticResource(v)

			default:
				obj[name] = v
			}

		case bool:
			if v {
				obj[name] = v
			}
		}
	}

	notification := make(map[string]interface{})
	notification["title"] = n.Title
	notification["target"] = n.Target
	setObjectField(notification, "lang", n.Lang)
	setObjectField(notification, "badge", n.Badge)
	setObjectField(notification, "body", n.Body)
	setObjectField(notification, "tag", n.Tag)
	setObjectField(notification, "icon", n.Icon)
	setObjectField(notification, "image", n.Image)
	setObjectField(notification, "data", n.Data)
	setObjectField(notification, "renotify", n.Renotify)
	setObjectField(notification, "requireInteraction", n.RequireInteraction)
	setObjectField(notification, "silent", n.Silent)

	if l := len(n.Vibrate); l != 0 {
		vibrate := make([]interface{}, l)
		for i, v := range n.Vibrate {
			vibrate[i] = v
		}
		notification["vibrate"] = vibrate
	}

	if l := len(n.Actions); l != 0 {
		actions := make([]interface{}, l)
		for i, a := range n.Actions {
			action := make(map[string]interface{}, 3)
			setObjectField(action, "action", a.Action)
			setObjectField(action, "title", a.Title)
			setObjectField(action, "icon", a.Icon)
			setObjectField(action, "target", a.Target)
			actions[i] = action
		}
		notification["actions"] = actions
	}

	Window().Call("goappNewNotification", notification)
}

func (ctx uiContext) cryptoKey() string {
	return strings.ReplaceAll(ctx.DeviceID(), "-", "")
}

func makeContext(src UI) Context {
	return uiContext{
		Context:            src.context(),
		src:                src,
		jsSrc:              src.JSValue(),
		appUpdateAvailable: appUpdateAvailable,
		page:               src.dispatcher().currentPage(),
		disp:               src.dispatcher(),
	}
}
