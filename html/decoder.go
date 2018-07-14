package html

import (
	"io"
	"strings"

	"github.com/murlokswarm/app"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Decoder is a tag decoder based on HTML5.
// It implements the app.TagDecoder interface.
type Decoder struct {
	tokenizer   *html.Tokenizer
	decodingSvg bool
	err         error
}

// NewDecoder create a tag decoder that reads from the given reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		tokenizer: html.NewTokenizer(r),
	}
}

// Decode decodes HTML5 and store it in the given tag.
// It satisfies the app.TagDecoder interface.
func (d *Decoder) Decode(tag *app.Tag) error {
	d.decode(tag)
	if d.err == io.EOF {
		d.err = nil
	}
	if d.err != nil {
		return d.err
	}

	if tag.Is(app.ZeroTag) {
		return errors.Errorf("no html to decode")
	}
	return nil
}

func (d *Decoder) decode(tag *app.Tag) bool {
	switch d.tokenizer.Next() {
	case html.StartTagToken:
		return d.decodeTag(tag)
	case html.EndTagToken:
		return d.decodeClosingTag(tag)
	case html.SelfClosingTagToken:
		return d.decodeSelfClosingTag(tag)
	case html.TextToken:
		return d.decodeText(tag)
	case html.ErrorToken:
		d.err = d.tokenizer.Err()
		return false
	default:
		return d.decode(tag)
	}
}

func (d *Decoder) decodeTag(tag *app.Tag) bool {
	name, hasAttr := d.tokenizer.TagName()
	tag.Name = string(name)
	tag.Type = app.SimpleTag
	tag.Svg = d.decodingSvg

	if hasAttr {
		d.decodeAttributes(tag)
	}

	switch {
	case tag.Name == "svg":
		d.decodingSvg = true
		tag.Svg = true
	case isVoidElement(tag.Name, d.decodingSvg):
		return true
	case isCompo(tag.Name, d.decodingSvg):
		tag.Type = app.CompoTag
		return true
	}

	for {
		var child app.Tag
		if !d.decode(&child) {
			return false
		}

		// A zero tag results from decoding a closing tag. It means there is no
		// more child to decode for this tag.
		if child.Is(app.ZeroTag) {
			return true
		}

		tag.Children = append(tag.Children, child)
	}
}

func (d *Decoder) decodeAttributes(tag *app.Tag) {
	attrs := make(app.AttributeMap)
	for {
		name, val, moreAttr := d.tokenizer.TagAttr()
		attrs[string(name)] = string(val)

		if !moreAttr {
			break
		}
	}
	tag.Attributes = attrs
}

func (d *Decoder) decodeClosingTag(tag *app.Tag) bool {
	name, _ := d.tokenizer.TagName()
	if string(name) == "svg" {
		d.decodingSvg = false
	}
	return true
}

func (d *Decoder) decodeSelfClosingTag(tag *app.Tag) bool {
	if !d.decodingSvg {
		d.err = errors.Errorf("decoding a self closing tag is not allowed outside a svg context")
		return false
	}

	name, hasAttr := d.tokenizer.TagName()
	tag.Name = string(name)
	tag.Type = app.SimpleTag
	tag.Svg = true

	if hasAttr {
		d.decodeAttributes(tag)
	}
	return true
}

func (d *Decoder) decodeText(tag *app.Tag) bool {
	tag.Text = string(d.tokenizer.Text())
	tag.Text = strings.TrimSpace(tag.Text)

	if len(tag.Text) == 0 {
		return d.decode(tag)
	}

	tag.Type = app.TextTag
	return true
}

func isVoidElement(name string, decodingSvg bool) bool {
	if decodingSvg {
		return false
	}
	_, ok := voidElems[name]
	return ok
}

var (
	voidElems = map[string]struct{}{
		"area":   {},
		"base":   {},
		"br":     {},
		"col":    {},
		"embed":  {},
		"hr":     {},
		"img":    {},
		"input":  {},
		"keygen": {},
		"link":   {},
		"meta":   {},
		"param":  {},
		"source": {},
		"track":  {},
		"wbr":    {},
	}
)

func isCompo(name string, decodingSvg bool) bool {
	if len(name) == 0 {
		return false
	}
	if decodingSvg {
		return false
	}

	// Any non standard html tag name describes a component name.
	return atom.Lookup([]byte(name)) == 0
}
