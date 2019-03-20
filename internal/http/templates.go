package http

// Code generated by go generate; DO NOT EDIT.

const pageHTML = `<!DOCTYPE html>
<html>

<head>
    <meta charset="UTF-8">
    <meta name="description" content="{{.Description}}">
    <meta name="keywords" content="{{.Keywords}}">
    <meta name="author" content="{{.Author}}">
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=0">
    <title>{{.Name}}</title>

    <style media="all" type="text/css">
{{.DefaultCSS}}
    </style>
    {{range .CSS}}<link type="text/css" rel="stylesheet" href="{{.}}">
    {{end}}
    <link rel="icon" type="image/png" href="/icon-192.png">
    <link rel="manifest" href="/manifest.json">
    {{range .Scripts}}<script src="{{.}}"></script>
    {{end}}
    <script>
{{.AppJS}}
    </script>
</head>
<body>
    <div class="App_Loader">
        <img id="App_LoadingIcon" class="App_InfiniteSpin" src="/icon-512.png">
        <p id="App_LoadingLabel">{{.LoadingLabel}}</p>
    </div>
</body>
</html>`

const pageCSS = `html {
    height: 100%;
    width: 100%;
    margin: 0;
    padding: 0;
    overflow: hidden;
}

body {
    height: 100%;
    width: 100%;
    margin: 0;
    padding: 0;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
    font-size: 10pt;
    font-weight: 400;
    color: white;
    background-color: #21252b;
}

h1 {
    margin: 0;
    padding: 30px 24px 6px;
    outline: 0;
    font-weight: 200;
    text-transform: lowercase;
    letter-spacing: 1px;
}

h2,
h3,
h4,
h5,
h6 {
    margin: 0;
    padding: 3px 24px 6px;
    outline: 0;
    font-weight: 200;
    text-transform: lowercase;
}

p {
    margin: 0;
    padding: 3px 24px 6px;
    outline: 0;
}

a {
    color: currentColor;
    text-decoration: none;
    cursor: pointer;
}

a:hover {
    color: deepskyblue;
}

ul {
    margin: 0;
    padding: 3px 24px 6px 42px;
    outline: 0;
}

ul li {
    margin: 0;
    padding: 3px 0;
}

ul li:first-child {
    padding: 0 0 3px;
}

ul li:last-child {
    padding: 3px 0 0;
}

table {
    width: calc(100% - 48px);
    margin: 0 24px;
    padding: 3px 0 6px;
    border-collapse: collapse;
    table-layout: fixed;
}

table th {
    padding: 0 12px 12px;
    border-bottom: 1px solid darkgray;
    font-size: 11pt;
    font-weight: bold;
    /* text-transform: lowercase; */
}

table td {
    padding: 12px;
    border-bottom: 0.1px solid darkgray;
    text-align: center;
}

table tr:first-child td {
    padding: 0 12px 12px;
}

table tr:last-child td {
    border-bottom: 0;
    padding: 12px;
}

button {
    background: none;
    border: 0;
    color: inherit;
    font: inherit;
    font-size: inherit;
    outline: inherit;
    -webkit-touch-callout: none;
    -webkit-user-select: none;
    -khtml-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
}

::-webkit-scrollbar {
    background-color: rgba(255, 255, 255, 0.04);
}

::-webkit-scrollbar-thumb {
    background-color: rgba(255, 255, 255, 0.05);
}

#App_ContextMenuBackground {
    display: none;
    position: fixed;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background-color: transparent;
}

#App_ContextMenu {
    position: absolute;
    min-width: 150px;
    max-width: 480px;
    padding: 6px 0;
    border-radius: 6px;
    border: solid 1px rgba(255, 255, 255, 0.1);
    background: #21252a;
    color: white;
    -webkit-box-shadow: -1px 12px 38px 0px rgba(0, 0, 0, 0.6);
    -moz-box-shadow: -1px 12px 38px 0px rgba(0, 0, 0, 0.6);
    box-shadow: -1px 12px 38px 0px rgba(0, 0, 0, 0.6);
}

.App_MenuItemSeparator {
    width: 100%;
    height: 0;
    margin: 6px 0;
    border-top: solid 1px rgba(255, 255, 255, 0.1);
}

.App_MenuItem {
    display: flex;
    padding: 3px 24px;
    width: 100%;
    text-align: left;
}

.App_MenuItem:disabled {
    opacity: 0.15;
    background-color: transparent;
}

.App_MenuItemLabel {
    user-select: none;
    flex-grow: 1;
}

.App_MenuItemKeys {
    flex-grow: 0;
    margin-left: 12px;
    text-transform: capitalize;
}

.App_MenuItem:hover {
    background-color: deepskyblue;
}

.App_Loader {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    width: 100%;
    height: 100%;
}

#App_LoadingIcon {
    max-width: 100px;
    max-height: 100px;
    user-select: none;
    -moz-user-select: none;
    -webkit-user-drag: none;
    -webkit-user-select: none;
    -ms-user-select: none;
    user-drag: none;
}

#App_LoadingLabel {
    margin-top: 6px;
    font-size: 16pt;
    font-weight: 100;
    text-transform: lowercase;
    letter-spacing: 1px;
    max-width: 480px;
    text-align: center;
}

.App_InfiniteSpin {
    animation: App_InfiniteSpinFrames 10s infinite linear;
}

@keyframes App_InfiniteSpinFrames {
    from {
        transform: rotate(0deg);
    }
    to {
        transform: rotate(359deg);
    }
}`

const pageJS = `// -----------------------------------------------------------------------------
// Goapp
// -----------------------------------------------------------------------------
var goapp = {
  nodes: {},

  actions: Object.freeze({
    'setRoot': 0,
    'newNode': 1,
    'delNode': 2,
    'setAttr': 3,
    'delAttr': 4,
    'setText': 5,
    'appendChild': 6,
    'removeChild': 7,
    'replaceChild': 8
  }),

  pointer: {
    x: 0,
    y: 0
  }
}

function render (changes = []) {
  changes.forEach(c => {
    switch (c.Action) {
      case goapp.actions.setRoot:
        setRoot(c)
        break

      case goapp.actions.newNode:
        newNode(c)
        break

      case goapp.actions.delNode:
        delNode(c)
        break

      case goapp.actions.setAttr:
        setAttr(c)
        break

      case goapp.actions.delAttr:
        delAttr(c)
        break

      case goapp.actions.setText:
        setText(c)
        break

      case goapp.actions.appendChild:
        appendChild(c)
        break

      case goapp.actions.removeChild:
        removeChild(c)
        break

      case goapp.actions.replaceChild:
        replaceChild(c)
        break

      default:
        console.log(c.Type + ' change is not supported')
    }
  })
}

function setRoot (change = {}) {
  const { NodeID } = change

  const n = goapp.nodes[NodeID]
  n.IsRootCompo = true

  const root = compoRoot(n)
  if (!root) {
    return
  }

  document.body.replaceChild(root, document.body.firstChild)
}

function newNode (change = {}) {
  const { IsCompo = false, Type, NodeID, CompoID, Namespace } = change

  if (IsCompo) {
    goapp.nodes[NodeID] = {
      Type,
      ID: NodeID,
      IsCompo
    }

    return
  }

  var n = null

  if (Type === 'text') {
    n = document.createTextNode('')
  } else if (change.Namespace) {
    n = document.createElementNS(Namespace, Type)
  } else {
    n = document.createElement(Type)
  }

  n.ID = NodeID
  n.CompoID = CompoID
  goapp.nodes[NodeID] = n
}

function delNode (change = {}) {
  const { NodeID } = change
  delete goapp.nodes[NodeID]
}

function setAttr (change = {}) {
  const { NodeID, Key, Value = '' } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  n.setAttribute(Key, Value)
}

function delAttr (change = {}) {
  const { NodeID, Key } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  n.removeAttribute(Key)
}

function setText (change = {}) {
  const { NodeID, Value } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  n.nodeValue = Value
}

function appendChild (change = {}) {
  const { NodeID, ChildID } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  if (n.IsCompo) {
    n.RootID = ChildID
    return
  }

  const c = compoRoot(goapp.nodes[ChildID])
  if (!c) {
    return
  }

  n.appendChild(c)
}

function removeChild (change = {}) {
  const { NodeID, ChildID } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  const c = compoRoot(goapp.nodes[ChildID])
  if (!c) {
    return
  }

  n.removeChild(c)
}

function replaceChild (change = {}) {
  const { NodeID, ChildID, NewChildID } = change

  const n = goapp.nodes[NodeID]
  if (!n) {
    return
  }

  const c = compoRoot(goapp.nodes[ChildID])
  if (!c) {
    return
  }

  const nc = compoRoot(goapp.nodes[NewChildID])
  if (!nc) {
    return
  }

  if (n.IsCompo) {
    n.RootID = NewChildID

    if (n.IsRootCompo) {
      setRoot({ NodeID: n.ID })
    }

    return
  }

  n.replaceChild(nc, c)
}

function compoRoot (node) {
  if (!node || !node.IsCompo) {
    return node
  }

  const n = goapp.nodes[node.RootID]
  return compoRoot(n)
}

function mapObject (obj) {
  var map = {}

  for (var field in obj) {
    const name = field[0].toUpperCase() + field.slice(1)
    const value = obj[field]
    const type = typeof value

    switch (type) {
      case 'object':
        break

      case 'function':
        break

      default:
        map[name] = value
        break
    }
  }

  return map
}

function callCompoHandler (elem, event, fieldOrMethod) {
  switch (event.type) {
    case 'change':
      onchangeToGolang(elem, fieldOrMethod)
      break

    case 'drag':
    case 'dragstart':
    case 'dragend':
    case 'dragexit':
      onDragStartToGolang(elem, event, fieldOrMethod)
      break

    case 'dragenter':
    case 'dragleave':
    case 'dragover':
    case 'drop':
      ondropToGolang(elem, event, fieldOrMethod)
      break

    case 'contextmenu':
      event.preventDefault()
      eventToGolang(elem, event, fieldOrMethod)
      trackPointerPosition(event)
      break

    default:
      eventToGolang(elem, event, fieldOrMethod)
      trackPointerPosition(event)
  }
}

function onchangeToGolang (elem, fieldOrMethod) {
  goapp.emit(JSON.stringify({
    'CompoID': elem.CompoID,
    'FieldOrMethod': fieldOrMethod,
    'JSONValue': JSON.stringify(elem.value)
  }))
}

function onDragStartToGolang (elem, event, fieldOrMethod) {
  const payload = mapObject(event.dataTransfer)
  payload['Data'] = elem.dataset.drag
  setPayloadSource(payload, elem)

  event.dataTransfer.setData('text', elem.dataset.drag)

  goapp.emit(JSON.stringify({
    'CompoID': elem.CompoID,
    'FieldOrMethod': fieldOrMethod,
    'JSONValue': JSON.stringify(payload)
  }))
}

function ondropToGolang (elem, event, fieldOrMethod) {
  event.preventDefault()

  const payload = mapObject(event.dataTransfer)
  payload['Data'] = event.dataTransfer.getData('text')
  payload['FileOverride'] = 'xxx'
  setPayloadSource(payload, elem)

  goapp.emit(JSON.stringify({
    'CompoID': elem.CompoID,
    'FieldOrMethod': fieldOrMethod,
    'JSONValue': JSON.stringify(payload),
    'Override': 'Files'
  }))
}

function eventToGolang (elem, event, fieldOrMethod) {
  const payload = mapObject(event)
  setPayloadSource(payload, elem)

  if (elem.contentEditable === 'true') {
    payload['InnerText'] = elem.innerText
  }

  goapp.emit(JSON.stringify({
    'CompoID': elem.CompoID,
    'FieldOrMethod': fieldOrMethod,
    'JSONValue': JSON.stringify(payload)
  }))
}

function setPayloadSource (payload, elem) {
  payload['Source'] = {
    'GoappID': elem.ID,
    'CompoID': elem.CompoID,
    'ID': elem.id,
    'Class': elem.className,
    'Data': elem.dataset,
    'Value': elem.value
  }
}

function trackPointerPosition (event) {
  if (event.clientX != undefined) {
    goapp.pointer.x = event.clientX
  }

  if (event.clientY != undefined) {
    goapp.pointer.y = event.clientY
  }
}

// -----------------------------------------------------------------------------
// Context menu
// -----------------------------------------------------------------------------

function showContextMenu () {
  const bg = document.getElementById('App_ContextMenuBackground')
  if (!bg) {
    console.log('no context menu declared')
    return
  }
  bg.style.display = 'block'

  const menu = document.getElementById('App_ContextMenu')

  const width = window.innerWidth ||
    document.documentElement.clientWidth ||
    document.body.clientWidth

  const height = window.innerHeight ||
    document.documentElement.clientHeight ||
    document.body.clientHeight

  var x = goapp.pointer.x
  if (x + menu.offsetWidth > width) {
    x = width - menu.offsetWidth - 1
  }

  var y = goapp.pointer.y
  if (y + menu.offsetHeight > height) {
    y = height - menu.offsetHeight - 1
  }

  menu.style.left = x + 'px'
  menu.style.top = y + 'px'
}

function hideContextMenu () {
  const bg = document.getElementById('App_ContextMenuBackground')
  if (!bg) {
    console.log('no context menu declared')
    return
  }
  bg.style.display = 'none'
}

// -----------------------------------------------------------------------------
// Init service worker
// -----------------------------------------------------------------------------
if ('serviceWorker' in navigator) {
  navigator.serviceWorker
    .register('/goapp.js')
    .then(reg => {
      console.log('offline service worker registered')
    })
    .catch(err => {
      console.error('offline service worker registration failed', err)
    })
}

// -----------------------------------------------------------------------------
// Init progressive app
// -----------------------------------------------------------------------------
let deferredPrompt

window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault()
  deferredPrompt = e
  console.log('beforeinstallprompt')
})

// -----------------------------------------------------------------------------
// Init Web Assembly
// -----------------------------------------------------------------------------
if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer()
    return await WebAssembly.instantiate(source, importObject)
  }
}

const go = new Go()

WebAssembly
  .instantiateStreaming(fetch('/goapp.wasm'), go.importObject)
  .then((result) => {
    go.run(result.instance)
  })
  .catch(err => {
    const loadingIcon = document.getElementById('App_LoadingIcon')
    loadingIcon.className = ''

    const loadingLabel = document.getElementById('App_LoadingLabel')
    loadingLabel.innerText = err
    console.error('wasm run failed: ' + err)
  })
`

const manifestJSON = `{
    "short_name": "{{.ShortName}}",
    "name": "{{.Name}}",
    "icons": [
        {
            "src": "/icon-192.png",
            "type": "image/png",
            "sizes": "192x192"
        },
        {
            "src": "/icon-512.png",
            "type": "image/png",
            "sizes": "512x512"
        }
    ],
    "start_url": "{{.StartURL}}",
    "background_color": "{{.BackgroundColor}}",
    "display": "{{.Display}}",
    "scope": "{{.Scope}}",
    "orientation": "{{.Orientation}}",
    "theme_color": "{{.ThemeColor}}"
}`
