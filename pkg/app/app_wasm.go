package app

import (
	"fmt"
	"net/url"
	"syscall/js"
)

var (
	window                = &browserWindow{value: value{Value: js.Global()}}
	body                  = Body()
	content     ValueNode = Div()
	contextMenu           = &contextMenuLayout{}
)

func run() {
	initContent()
	initContextMenu()

	LocalStorage = newJSStorage("localStorage")
	SessionStorage = newJSStorage("sessionStorage")

	onnav := FuncOf(onNavigate)
	defer onnav.Release()
	Window().Set("onclick", onnav)

	onpopstate := FuncOf(onPopState)
	defer onpopstate.Release()
	Window().Set("onpopstate", onpopstate)

	url := Window().URL()

	if err := navigate(url, false); err != nil {
		panic(fmt.Errorf("navigating to %s failed: %w", url, err))
	}

	select {}
}

func initContent() {
	body.value = Window().Get("document").Get("body")
	content.(*HTMLDiv).value = body.value.Get("firstElementChild")
	content.setParent(body)
	body.appendChild(content)
}

func initContextMenu() {
	rawContextMenu := Div().ID("app-context-menu")
	rawContextMenu.value = Window().
		Get("document").
		Call("getElementById", "app-context-menu")
	rawContextMenu.setParent(body)
	body.appendChild(rawContextMenu)
	if err := update(rawContextMenu, contextMenu); err != nil {
		panic(fmt.Errorf("initializing context menu failed: %w", err))
	}
}

func onNavigate(this Value, args []Value) interface{} {
	event := args[0]
	elem := event.Get("target")
	if !elem.Truthy() {
		elem = event.Get("srcElement")
	}

	var u string
	switch elem.Get("tagName").String() {
	case "A":
		u = elem.Get("href").String()

	default:
		return nil
	}

	event.Call("preventDefault")
	Navigate(u)
	return nil

}

func onPopState(this Value, args []Value) interface{} {
	if u := Window().URL(); u.Fragment == "" {
		navigate(Window().URL(), false)
	}
	return nil
}

func navigate(u *url.URL, updateHistory bool) error {
	if !isPWANavigation(u) {
		Window().Get("location").Set("href", u.String())
		return nil
	}

	root, ok := routes[u.Path]
	if !ok {
		root = NotFound
	}
	if content == root {
		return nil
	}
	if err := replace(content, root); err != nil {
		return err
	}
	content = root

	if updateHistory {
		Window().Get("history").Call("pushState", nil, "", u.String())
	}

	return nil
}

func isPWANavigation(u *url.URL) bool {
	externalNav := u.Host != "" && u.Host != Window().URL().Host
	fragmentNav := u.Fragment != ""
	return !externalNav && !fragmentNav
}

func reload() {
	Window().Get("location").Call("reload")
}

func newContextMenu(menuItems ...MenuItemNode) {
	contextMenu.show(menuItems...)
}