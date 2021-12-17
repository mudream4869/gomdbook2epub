// This extension adds replacer renderer to change urls.

package srcreplacer

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// ReplaceFunc is a function for replacing source link.
type ReplaceFunc = func(link string) string

type withReplacer struct {
	value *replacer
}

func (o *withReplacer) SetConfig(c *renderer.Config) {
	c.NodeRenderers = append(c.NodeRenderers, util.Prioritized(o.value, 0))
}

// replacer render image and link with replaced source link.
type replacer struct {
	html.Config
	imgReplacer  ReplaceFunc
	linkReplacer ReplaceFunc
}

// New return initialized renderer with source url replacing support.
func New(imgReplacer, linkReplacer ReplaceFunc, options ...html.Option) goldmark.Extender {
	var config = html.NewConfig()
	for _, opt := range options {
		opt.SetHTMLOption(&config)
	}
	return &replacer{
		Config:       config,
		imgReplacer:  imgReplacer,
		linkReplacer: linkReplacer,
	}
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs interface.
func (r *replacer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
}

func (r *replacer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)

	// add image source replacing hack
	if r.imgReplacer != nil {
		var src = r.imgReplacer(util.BytesToReadOnlyString(n.Destination))
		n.Destination = util.StringToReadOnlyBytes(src)
	}

	w.WriteString("<img src=\"")
	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}
	w.WriteString(`" alt="`)
	w.Write(n.Text(source))
	w.WriteByte('"')
	if n.Title != nil {
		w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		w.WriteByte('"')
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	if r.XHTML {
		w.WriteString(" />")
	} else {
		w.WriteString(">")
	}
	return ast.WalkSkipChildren, nil
}

func (r *replacer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		// add link source replacing hack
		if r.linkReplacer != nil {
			var src = r.linkReplacer(util.BytesToReadOnlyString(n.Destination))
			n.Destination = util.StringToReadOnlyBytes(src)
		}

		_, _ = w.WriteString("<a href=\"")
		if r.Unsafe || !html.IsDangerousURL(n.Destination) {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		}
		_ = w.WriteByte('"')
		if n.Title != nil {
			_, _ = w.WriteString(` title="`)
			r.Writer.Write(w, n.Title)
			_ = w.WriteByte('"')
		}
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.LinkAttributeFilter)
		}
		_ = w.WriteByte('>')
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// Extend implement goldmark.Extender interface.
func (r *replacer) Extend(m goldmark.Markdown) {
	if r.imgReplacer == nil && r.linkReplacer == nil {
		return
	}

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(r, 0),
		),
	)
}
