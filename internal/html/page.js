
var goapp = {
    nodes: {}
};

function render(changes = []) {
    changes.forEach(c => {
        switch (c.Type) {
            case 'createText':
                break;
            case 'setText':
                break;
            case 'createElem':
                break;
            case 'setAttrs':
                break;
            case 'appendChild':
                break;
            case 'removeChild':
                break;
            case 'replaceChild':
                break;
            case 'createCompo':
                break;
            case 'setCompoRoot':
                break;
            case 'deleteNode':
                break;
            default:
                console.log('unknown change: ' + c.Type);
        }
    });
}

// render replaces the attributes of the node with the given id by the given
// attributes.
// function renderAttributes(payload) {
//     const { id, attributes } = payload;

//     if (!attributes) {
//         return;
//     }

//     const selector = '[data-goapp-id="' + id + '"]';
//     const elem = document.querySelector(selector);

//     if (!elem) {
//         return;
//     }

//     if (!elem.hasAttributes()) {
//         return;
//     }
//     const elemAttrs = elem.attributes;

//     // Remove missing attributes.
//     for (var i = 0; i < elemAttrs.length; i++) {
//         const name = elemAttrs[i].name;

//         if (name === 'data-goapp-id') {
//             continue;
//         }

//         if (attributes[name] === undefined) {
//             elem.removeAttribute(name);
//         }
//     }

//     // Set attributes.
//     for (var name in attributes) {
//         const currentValue = elem.getAttribute(name);
//         const newValue = attributes[name];

//         if (name === 'value') {
//             elem.value = newValue;
//             continue;
//         }

//         if (currentValue !== newValue) {
//             elem.setAttribute(name, newValue);
//         }
//     }
// }

function mapObject(obj) {
    var map = {};

    for (var field in obj) {
        const name = field[0].toUpperCase() + field.slice(1);
        const value = obj[field];
        const type = typeof value;

        switch (type) {
            case 'object':
                break;

            case 'function':
                break;

            default:
                map[name] = value;
                break;
        }
    }
    return map;
}

function callCompoHandler(compoID, target, src, event) {
    var payload = null;

    switch (event.type) {
        case 'change':
            onchangeToGolang(compoID, target, src, event);
            break;

        case 'drag':
        case 'dragstart':
        case 'dragend':
        case 'dragexit':
            onDragStartToGolang(compoID, target, src, event);
            break;

        case 'dragenter':
        case 'dragleave':
        case 'dragover':
        case 'drop':
            ondropToGolang(compoID, target, src, event);
            break;

        default:
            eventToGolang(compoID, target, src, event);
            break;
    }
}

function onchangeToGolang(compoID, target, src, event) {
    golangRequest(JSON.stringify({
        'compo-id': compoID,
        'target': target,
        'json-value': JSON.stringify(src.value)
    }));
}

function onDragStartToGolang(compoID, target, src, event) {
    const payload = mapObject(event.dataTransfer);
    payload['Data'] = src.dataset.drag;

    event.dataTransfer.setData('text', src.dataset.drag);

    golangRequest(JSON.stringify({
        'compo-id': compoID,
        'target': target,
        'json-value': JSON.stringify(payload)
    }));
}

function ondropToGolang(compoID, target, src, event) {
    event.preventDefault();

    const payload = mapObject(event.dataTransfer);
    payload['Data'] = event.dataTransfer.getData('text');
    payload['file-override'] = 'xxx';

    golangRequest(JSON.stringify({
        'compo-id': compoID,
        'target': target,
        'json-value': JSON.stringify(payload),
        'override': 'Files'
    }));
}

function eventToGolang(compoID, target, src, event) {
    const payload = mapObject(event);

    if (src.contentEditable === 'true') {
        payload['InnerText'] = src.innerText;
    }

    golangRequest(JSON.stringify({
        'compo-id': compoID,
        'target': target,
        'json-value': JSON.stringify(payload)
    }));
}
