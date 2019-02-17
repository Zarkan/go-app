// +build wasm

package app

import (
	"strings"
	"syscall/js"
)

func init() {
	Import(&ContextMenu{})
}

// ContextMenu is a component that displays a context menu.
type ContextMenu struct {
	Items   []MenuItem
	Visible bool
}

// Subscribe set up the context menu listened events.
func (m *ContextMenu) Subscribe() *Subscriber {
	return NewSubscriber().Subscribe("app.NewContextMenu", m.new)
}

// Render returns the markup that describes the context menu.
func (m *ContextMenu) Render() string {
	return `
<div id="App_ContextMenuBackground" onclick="Hide">
	<div id="App_ContextMenu">
		{{range $idx, $v := .Items}}
	
		{{if .Separator}}
		<div class="App_MenuItemSeparator"></div>
		{{else}}
		<button class="App_MenuItem" onclick="{{bind "Items" $idx "OnClick"}}" {{if .Disabled}}disabled{{end}}>
			<div class="App_MenuItemLabel">{{.Label}}</div>
			<div class="App_MenuItemKeys">{{.Keys}}</div>
		</button>
		{{end}}

		{{end}}
	</div>
</div>
	`
}

func (m *ContextMenu) new(items []MenuItem) {
	for i := range items {
		if items[i].OnClick == nil {
			items[i].Disabled = true
		}

		items[i].Keys = convertKeys(items[i].Keys)
	}

	m.Items = items
	m.Visible = true
	Render(m)

	UI(func() {
		js.Global().Call("showContextMenu")
	})
}

func (m *ContextMenu) Hide() {
	js.Global().Call("hideContextMenu")
}

func convertKeys(k string) string {
	k = strings.ToLower(k)

	switch js.Global().Get("navigator").Get("platform").String() {
	case "Macintosh", "MacIntel", "MacPPC", "Mac68K":
		k = strings.Replace(k, "cmdorctrl", "⌘", -1)
		k = strings.Replace(k, "cmd", "⌘", -1)
		k = strings.Replace(k, "command", "⌘", -1)
		k = strings.Replace(k, "ctrl", "⌃", -1)
		k = strings.Replace(k, "control", "⌃", -1)
		k = strings.Replace(k, "alt", "⌥", -1)
		k = strings.Replace(k, "option", "⌥", -1)
		k = strings.Replace(k, "shift", "⇧", -1)
		k = strings.Replace(k, "capslock", "⇪", -1)
		k = strings.Replace(k, "del", "⌫", -1)
		k = strings.Replace(k, "delete", "⌫", -1)
		k = strings.Replace(k, "+", "", -1)

	case "Windows", "Win32":
		k = strings.Replace(k, "cmdorctrl", "ctrl", -1)
		k = strings.Replace(k, "cmd", "win", -1)
		k = strings.Replace(k, "command", "win", -1)
		k = strings.Replace(k, "control", "ctrl", -1)

	default:
		k = ""
	}

	return k
}
