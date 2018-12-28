package dom

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"sync"

	"github.com/murlokswarm/app"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// Engine represents a dom (document object model) engine.
// It manages components an nodes lifecycle and keep track of node changes.
// The engine can be synchronized with a remote dom like a web browser document.
// It is safe for concurrent operations.
type Engine struct {
	// The factory to decode component from html.
	Factory *app.Factory

	// The func that generates a resources directory path usable in html.
	Resources func(...string) string

	// AttrTransforms describes a set of transformation to apply to parsed node
	// attributes.
	AttrTransforms []Transform

	// AllowedNodes define the allowed node type. All node type is allowed when
	// the slice is empty.
	AllowedNodes []string

	// Sync is the function used to synchronize node changes with a remote dom.
	// No synchronisations are performed if the func in nil.
	Sync func(arg interface{}) error

	// UI the function to execute the given func on the ui goroutine.
	UI func(func())

	once          sync.Once
	mutex         sync.RWMutex
	compos        map[app.Compo]compo
	compoIDs      map[string]compo
	nodes         map[string]node
	allowdedNodes map[string]struct{}
	rootID        string
	creates       []change
	changes       []change
	deletes       []change
	toSync        []change
	decodeAttrs   map[string]string
}

func (e *Engine) init() {
	if e.Sync == nil {
		e.Sync = func(v interface{}) error {
			return nil
		}
	}

	e.compos = make(map[app.Compo]compo)
	e.compoIDs = make(map[string]compo)
	e.nodes = make(map[string]node)

	if len(e.AllowedNodes) != 0 {
		e.allowdedNodes = make(map[string]struct{}, len(e.AllowedNodes))

		for _, a := range e.AllowedNodes {
			e.allowdedNodes[a] = struct{}{}
		}
	}

	if e.UI == nil {
		e.UI = func(f func()) {}
	}

	e.creates = make([]change, 0, 64)
	e.changes = make([]change, 0, 64)
	e.deletes = make([]change, 0, 64)
	e.toSync = make([]change, 0, 64)

	e.decodeAttrs = make(map[string]string)
}

// Contains reports whether the given component is in the dom.
func (e *Engine) Contains(c app.Compo) bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	_, ok := e.compos[c]
	return ok
}

// CompoByID returns the component with the given identifier.
func (e *Engine) CompoByID(id string) (app.Compo, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	c, ok := e.compoIDs[id]
	if !ok {
		return nil, app.ErrCompoNotMounted
	}

	return c.Compo, nil
}

// New renders the given component and set it as the dom root.
func (e *Engine) New(c app.Compo) error {
	e.once.Do(e.init)
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.close()

	if err := e.render(c); err != nil {
		return err
	}

	ic := e.compos[c]
	e.rootID = ic.ID

	e.changes = append(e.changes, change{
		Action: setRoot,
		NodeID: ic.ID,
	})

	return e.sync()
}

// Close deletes the components and nodes from the dom.
func (e *Engine) Close() {
	e.once.Do(e.init)
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.close()

}

func (e *Engine) close() {
	e.deleteNode(e.rootID)
	e.rootID = ""

	for k := range e.compos {
		delete(e.compos, k)
	}

	for k := range e.compoIDs {
		delete(e.compoIDs, k)
	}

	for k := range e.nodes {
		delete(e.nodes, k)
	}

	e.creates = clearChanges(e.creates)
	e.changes = clearChanges(e.changes)
	e.deletes = clearChanges(e.deletes)
	e.toSync = clearChanges(e.toSync)
}

// Render renders the given component by updating the state described within
// c.Render().
func (e *Engine) Render(c app.Compo) error {
	e.once.Do(e.init)
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if _, ok := e.compos[c]; !ok {
		return app.ErrCompoNotMounted
	}

	if err := e.render(c); err != nil {
		return err
	}

	return e.sync()
}

func (e *Engine) render(c app.Compo) error {
	ic, ok := e.compos[c]
	if !ok {
		typ := app.CompoName(c)

		if err := e.newCompo(c, node{
			ID:       genNodeID(typ),
			Type:     typ,
			ChildIDs: make([]string, 1),
			IsCompo:  true,
		}); err != nil {
			return err
		}

		ic = e.compos[c]
	}

	n := e.nodes[ic.ID]
	root := node{}
	newRoot := node{}

	if len(n.ChildIDs) != 0 {
		root = e.nodes[n.ChildIDs[0]]
	}

	markup, err := e.compoToHTML(c)
	if err != nil {
		return errors.Wrap(err, "reading component failed")
	}

	if newRoot, _, err = e.renderNode(rendering{
		Tokenizer:  html.NewTokenizer(bytes.NewBufferString(markup)),
		CompoID:    n.ID,
		NodeToSync: root,
	}); err != nil {
		return err
	}

	n.ChildIDs[0] = newRoot.ID
	e.nodes[n.ID] = n

	switch {
	case len(root.ID) == 0:
		e.changes = append(e.changes, change{
			Action:  appendChild,
			NodeID:  n.ID,
			ChildID: newRoot.ID,
		})

	case root.ID != newRoot.ID:
		e.deleteNode(root.ID)
		e.changes = append(e.changes, change{
			Action:     replaceChild,
			NodeID:     n.ID,
			ChildID:    root.ID,
			NewChildID: newRoot.ID,
		})
	}

	return nil
}

func (e *Engine) compoToHTML(c app.Compo) (string, error) {
	var extendedFuncs map[string]interface{}
	if extended, ok := c.(app.CompoWithExtendedRender); ok {
		extendedFuncs = extended.Funcs()
	}

	// The number of template functions. It contains the
	// component extended functions, the converters and
	// the resources accessor.
	funcsCount := len(converters) + len(extendedFuncs) + 1

	funcs := make(template.FuncMap, funcsCount)
	funcs["resources"] = e.Resources

	for k, v := range converters {
		funcs[k] = v
	}

	for k, v := range extendedFuncs {
		if _, ok := funcs[k]; ok {
			return "", errors.Errorf("template extension can't be named %s", k)
		}
		funcs[k] = v
	}

	tmpl, err := template.
		New(fmt.Sprintf("%T", c)).
		Funcs(funcs).
		Parse(c.Render())
	if err != nil {
		return "", err
	}

	var w bytes.Buffer
	if err = tmpl.Execute(&w, c); err != nil {
		return "", err
	}

	html := strings.TrimSpace(w.String())
	if len(html) == 0 {
		return "", errors.New("component does not render anything")
	}

	return html, nil
}

func (e *Engine) renderNode(r rendering) (node, bool, error) {
	switch r.Tokenizer.Next() {
	case html.TextToken:
		return e.renderText(r)

	case html.SelfClosingTagToken:
		return e.renderSelfClosingTag(r)

	case html.StartTagToken:
		return e.renderStartTag(r)

	case html.EndTagToken:
		return node{}, false, nil

	case html.ErrorToken:
		return node{}, false, r.Tokenizer.Err()

	default:
		return e.renderNode(r)
	}
}

func (e *Engine) renderText(r rendering) (node, bool, error) {
	text := string(r.Tokenizer.Text())
	text = strings.TrimSpace(text)

	if len(text) == 0 || len(r.Namespace) != 0 {
		// Invalid text, iterator next node.
		return e.renderNode(r)
	}

	n := r.NodeToSync

	if len(r.NodeToSync.ID) == 0 || r.NodeToSync.Type != "text" {
		n = node{
			ID:      genNodeID("text"),
			CompoID: r.CompoID,
			Type:    "text",
			Dom:     e,
		}
		e.newNode(n)
	}

	if text != n.Text {
		n.Text = text
		e.changes = append(e.changes, change{
			Action: setText,
			NodeID: n.ID,
			Value:  text,
		})
		e.nodes[n.ID] = n
	}

	return n, true, nil
}

func (e *Engine) renderSelfClosingTag(r rendering) (node, bool, error) {
	tagName, hasAttr := r.Tokenizer.TagName()
	typ := string(tagName)

	if typ == "svg" {
		r.Namespace = svg
	}

	if isCompoNode(typ, r.Namespace) {
		return e.renderCompoNode(r, typ, hasAttr)
	}

	if !e.isAllowedNode(typ) {
		return node{}, false, errors.Errorf("%s is not allowed", typ)
	}

	n := r.NodeToSync

	if len(n.ID) == 0 || n.Type != typ {
		n = node{
			ID:        genNodeID(typ),
			CompoID:   r.CompoID,
			Type:      typ,
			Namespace: r.Namespace,
			Dom:       e,
		}
		e.newNode(n)
	}

	n = e.renderTagAttrs(r, n, hasAttr, true)

	for _, childID := range n.ChildIDs {
		e.deleteNode(childID)
		e.changes = append(e.changes, change{
			Action:  removeChild,
			NodeID:  n.ID,
			ChildID: childID,
		})
	}

	n.ChildIDs = clearNodeIDs(n.ChildIDs)
	e.nodes[n.ID] = n
	return n, true, nil
}

func (e *Engine) renderStartTag(r rendering) (node, bool, error) {
	tagName, hasAttr := r.Tokenizer.TagName()
	typ := string(tagName)

	if typ == "svg" {
		r.Namespace = svg
	}

	if isCompoNode(typ, r.Namespace) {
		return e.renderCompoNode(r, typ, hasAttr)
	}

	if !e.isAllowedNode(typ) {
		return node{}, false, errors.Errorf("%s is not allowed", typ)
	}

	n := r.NodeToSync

	if len(n.ID) == 0 || n.Type != typ {
		n = node{
			ID:        genNodeID(typ),
			CompoID:   r.CompoID,
			Type:      typ,
			Namespace: r.Namespace,
			Dom:       e,
		}
		e.newNode(n)
	}

	n = e.renderTagAttrs(r, n, hasAttr, true)

	if isVoidElem(n.Type) {
		return n, true, nil
	}

	childIDs := n.ChildIDs
	moreChild := true
	count := 0

	// Replace children:
	for len(childIDs) != 0 {
		var err error

		old := e.nodes[childIDs[0]]
		new := node{}

		new, moreChild, err = e.renderNode(rendering{
			Tokenizer:  r.Tokenizer,
			CompoID:    r.CompoID,
			Namespace:  r.Namespace,
			NodeToSync: old,
		})

		if err != nil {
			return node{}, false, err
		}

		if !moreChild {
			break
		}

		if new.ID != old.ID {
			e.changes = append(e.changes, change{
				Action:     replaceChild,
				NodeID:     n.ID,
				ChildID:    old.ID,
				NewChildID: new.ID,
			})

			childIDs[0] = new.ID
			new.ParentID = n.ID
			e.nodes[new.ID] = new
			e.deleteNode(old.ID)
		}

		count++
		childIDs = childIDs[1:]
	}

	// Remove children:
	for _, childID := range childIDs {
		e.deleteNode(childID)
		e.changes = append(e.changes, change{
			Action:  removeChild,
			NodeID:  n.ID,
			ChildID: childID,
		})

	}
	childIDs = clearNodesIDsFrom(n.ChildIDs, count)

	// Add children
	for moreChild {
		var child node
		var err error

		child, moreChild, err = e.renderNode(rendering{
			Tokenizer: r.Tokenizer,
			CompoID:   r.CompoID,
			Namespace: r.Namespace,
		})

		if err != nil {
			return node{}, false, err
		}

		if !moreChild {
			break
		}

		childIDs = append(childIDs, child.ID)
		e.changes = append(e.changes, change{
			Action:  appendChild,
			NodeID:  n.ID,
			ChildID: child.ID,
		})
	}

	n.ChildIDs = childIDs
	e.nodes[n.ID] = n
	return n, true, nil
}

func (e *Engine) renderTagAttrs(r rendering, n node, moreAttr, changes bool) node {
	if !moreAttr {
		return n
	}

	if len(n.Attrs) == 0 {
		n.Attrs = make(map[string]string)
	}

	for moreAttr {
		var rk []byte
		var rv []byte

		rk, rv, moreAttr = r.Tokenizer.TagAttr()
		v := string(rv)

		k := string(rk)
		if r.Namespace == svg {
			k = svgAttr(k)
		}

		for _, t := range e.AttrTransforms {
			k, v = t(k, v)
		}

		e.decodeAttrs[k] = v
		if currentVal, ok := n.Attrs[k]; ok && currentVal == v {
			continue
		}

		n.Attrs[k] = v

		if changes {
			e.changes = append(e.changes, change{
				Action: setAttr,
				NodeID: n.ID,
				Key:    k,
				Value:  v,
			})
		}
	}

	for k := range n.Attrs {
		if _, ok := e.decodeAttrs[k]; ok {
			continue
		}

		delete(n.Attrs, k)

		if changes {
			e.changes = append(e.changes, change{
				Action: delAttr,
				NodeID: n.ID,
				Key:    k,
			})
		}
	}

	for k := range e.decodeAttrs {
		delete(e.decodeAttrs, k)
	}

	e.nodes[n.ID] = n
	return n
}

func (e *Engine) renderCompoNode(r rendering, typ string, hasAttr bool) (node, bool, error) {
	n := r.NodeToSync

	if len(n.ID) == 0 || n.Type != typ {
		n = node{
			ID:       genNodeID(typ),
			CompoID:  r.CompoID,
			Type:     typ,
			ChildIDs: make([]string, 1),
			IsCompo:  true,
			Dom:      e,
		}

		if err := e.newCompo(nil, n); err != nil {
			return node{}, false, err
		}
	}

	e.nodes[n.ID] = n
	n = e.renderTagAttrs(r, n, hasAttr, false)
	c := e.compoIDs[n.ID]

	if err := mapCompoFields(c.Compo, n.Attrs); err != nil {
		return node{}, false, err
	}

	if err := e.render(c.Compo); err != nil {
		return n, false, errors.Wrapf(err, "rendering %s failed", n.Type)
	}

	return n, true, nil
}

func (e *Engine) newNode(n node) {
	e.nodes[n.ID] = n

	e.creates = append(e.creates, change{
		Action:    newNode,
		NodeID:    n.ID,
		CompoID:   n.CompoID,
		Type:      n.Type,
		Namespace: n.Namespace,
		IsCompo:   n.IsCompo,
	})
}

func (e *Engine) newCompo(c app.Compo, n node) error {
	var err error
	if c == nil {
		if c, err = e.Factory.NewCompo(n.Type); err != nil {
			return err
		}
	}

	if err := validateCompo(c); err != nil {
		return err
	}

	e.newNode(n)

	ic := compo{
		ID:    n.ID,
		Compo: c,
	}

	if sub, ok := c.(app.EventSubscriber); ok {
		ic.Events = sub.Subscribe()
	}

	e.compoIDs[n.ID] = ic
	e.compos[c] = ic

	if mounter, ok := c.(app.Mounter); ok {
		e.UI(mounter.OnMount)
	}

	return nil
}

func (e *Engine) deleteNode(id string) {
	n, ok := e.nodes[id]
	if !ok {
		return
	}

	for _, childID := range n.ChildIDs {
		e.deleteNode(childID)
	}

	if n.IsCompo {
		if c, ok := e.compoIDs[n.ID]; ok {
			if c.Events != nil {
				c.Events.Close()
			}

			if dismounter, ok := c.Compo.(app.Dismounter); ok {
				e.UI(dismounter.OnDismount)
			}

			delete(e.compos, c.Compo)
			delete(e.compoIDs, c.ID)
		}
	}

	delete(e.nodes, n.ID)
	e.deletes = append(e.deletes, change{
		Action: delNode,
		NodeID: n.ID,
	})
}

func (e *Engine) sync() error {
	e.toSync = append(e.toSync, e.creates...)
	e.toSync = append(e.toSync, e.changes...)
	e.toSync = append(e.toSync, e.deletes...)

	if err := e.Sync(e.toSync); err != nil {
		return errors.Wrap(err, "syncing dom failed")
	}

	e.creates = clearChanges(e.creates)
	e.changes = clearChanges(e.changes)
	e.deletes = clearChanges(e.deletes)
	e.toSync = clearChanges(e.toSync)
	return nil
}

func (e *Engine) isAllowedNode(typ string) bool {
	if len(e.allowdedNodes) == 0 {
		return true
	}

	_, ok := e.allowdedNodes[typ]
	return ok
}

func validateCompo(c app.Compo) error {
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr {
		return errors.New("compo is not a pointer")
	}

	v = v.Elem()
	if v.NumField() == 0 {
		return errors.New("compo is based on a struct without field. use app.ZeroCompo instead of struct{}")
	}
	return nil
}

type rendering struct {
	Tokenizer  *html.Tokenizer
	CompoID    string
	Namespace  string
	NodeToSync node
}
