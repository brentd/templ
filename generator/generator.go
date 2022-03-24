package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/parser"
)

func NewRangeWriter(w io.Writer) *RangeWriter {
	return &RangeWriter{
		Current: parser.NewPosition(),
		w:       w,
	}
}

type RangeWriter struct {
	Current parser.Position
	w       io.Writer
}

func (rw *RangeWriter) WriteIndent(level int, s string) (r parser.Range, err error) {
	_, err = rw.Write(strings.Repeat("\t", level))
	if err != nil {
		return
	}
	return rw.Write(s)
}

func (rw *RangeWriter) Write(s string) (r parser.Range, err error) {
	r.From = parser.Position{
		Index: rw.Current.Index,
		Line:  rw.Current.Line,
		Col:   rw.Current.Col,
	}
	var n int
	for _, c := range s {
		if c == '\n' {
			rw.Current.Line++
			rw.Current.Col = 0
		}
		rw.Current.Col++
		n, err = io.WriteString(rw.w, string(c))
		rw.Current.Index += int64(n)
		if err != nil {
			return r, err
		}
	}
	r.To = rw.Current
	return r, err
}

func Generate(template parser.TemplateFile, w io.Writer) (sm *parser.SourceMap, err error) {
	g := generator{
		tf:        template,
		w:         NewRangeWriter(w),
		sourceMap: parser.NewSourceMap(),
	}
	err = g.generate()
	sm = g.sourceMap
	return
}

type generator struct {
	tf         parser.TemplateFile
	w          *RangeWriter
	sourceMap  *parser.SourceMap
	variableID int
}

func (g *generator) generate() (err error) {
	if err = g.writeCodeGeneratedComment(); err != nil {
		return
	}
	if err = g.writePackage(); err != nil {
		return
	}
	if err = g.writeImports(); err != nil {
		return
	}
	if err = g.writeTemplateNodes(); err != nil {
		return
	}
	return err
}

// Binary builds set this version string. goreleaser sets the value using Go build ldflags.
var version string

// Source builds use this value. When installed using `go install github.com/a-h/templ/cmd/templ@latest` the `version` variable is empty, but
// the debug.ReadBuildInfo return value provides the package version number installed by `go install`
func goInstallVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

func getVersion() string {
	if version != "" {
		return version
	}
	return goInstallVersion()
}

func (g *generator) writeCodeGeneratedComment() error {
	_, err := g.w.Write(fmt.Sprintf("// Code generated by templ@%s DO NOT EDIT.\n\n", getVersion()))
	return err
}

func (g *generator) writePackage() error {
	var r parser.Range
	var err error
	// package
	if _, err = g.w.Write("package "); err != nil {
		return err
	}
	if r, err = g.w.Write(g.tf.Package.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(g.tf.Package.Expression, r)
	if _, err = g.w.Write("\n\n"); err != nil {
		return err
	}
	if _, err = g.w.Write("//lint:file-ignore SA4006 This context is only used if a nested component is present.\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) templateNodeInfo() (hasTemplates bool, hasCSS bool) {
	for _, n := range g.tf.Nodes {
		switch n.(type) {
		case parser.HTMLTemplate:
			hasTemplates = true
		case parser.CSSTemplate:
			hasCSS = true
		}
		if hasTemplates && hasCSS {
			return
		}
	}
	return
}

func (g *generator) writeImports() error {
	var r parser.Range
	var err error
	// Always import templ because it's the interface type of all templates.
	if _, err = g.w.Write("import \"github.com/a-h/templ\"\n"); err != nil {
		return err
	}
	hasTemplates, hasCSS := g.templateNodeInfo()
	if hasTemplates {
		// The first parameter of a template function.
		if _, err = g.w.Write("import \"context\"\n"); err != nil {
			return err
		}
		// The second parameter of a template function.
		if _, err = g.w.Write("import \"io\"\n"); err != nil {
			return err
		}
		// Used to buffer writes.
		if _, err = g.w.Write("import \"bufio\"\n"); err != nil {
			return err
		}
	}
	if hasCSS {
		// strings.Builder is used to create CSS.
		if _, err = g.w.Write("import \"strings\"\n"); err != nil {
			return err
		}
	}
	for _, im := range g.tf.Imports {
		// import
		if _, err = g.w.Write("import "); err != nil {
			return err
		}
		if r, err = g.w.Write(im.Expression.Value); err != nil {
			return err
		}
		g.sourceMap.Add(im.Expression, r)
		if _, err = g.w.Write("\n"); err != nil {
			return err
		}
	}
	if _, err = g.w.Write("\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplateNodes() error {
	for i := 0; i < len(g.tf.Nodes); i++ {
		switch n := g.tf.Nodes[i].(type) {
		case parser.HTMLTemplate:
			if err := g.writeTemplate(n); err != nil {
				return err
			}
		case parser.CSSTemplate:
			if err := g.writeCSS(n); err != nil {
				return err
			}
		case parser.ScriptTemplate:
			if err := g.writeScript(n); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown node type: %v", reflect.TypeOf(n))
		}
	}
	return nil
}

func (g *generator) writeCSS(n parser.CSSTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(n.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Name, r)
	// () templ.CSSClass {
	if _, err = g.w.Write("() templ.CSSClass {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// var templCSSBuilder strings.Builder
		if _, err = g.w.WriteIndent(indentLevel, "var templCSSBuilder strings.Builder\n"); err != nil {
			return err
		}
		for i := 0; i < len(n.Properties); i++ {
			switch p := n.Properties[i].(type) {
			case parser.ConstantCSSProperty:
				// Carry out sanitization at compile time for constants.
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSBuilder.WriteString(`%s`)\n", templ.SanitizeCSS(p.Name, p.Value))); err != nil {
					return err
				}
			case parser.ExpressionCSSProperty:
				// templCSSBuilder.WriteString(templ.SanitizeCSS('name', p.Expression()))
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSBuilder.WriteString(string(templ.SanitizeCSS(`%s`, ", p.Name)); err != nil {
					return err
				}
				if r, err = g.w.Write(p.Value.Expression.Value); err != nil {
					return err
				}
				g.sourceMap.Add(p.Value.Expression, r)
				if _, err = g.w.Write(")))\n"); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unknown CSS property type: %v", reflect.TypeOf(p))
			}
		}
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("templCSSID := templ.CSSID(`%s`, templCSSBuilder.String())\n", n.Name.Value)); err != nil {
			return err
		}
		// return templ.CSS {
		if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentCSSClass{\n"); err != nil {
			return err
		}
		{
			indentLevel++
			// ID: templCSSID,
			if _, err = g.w.WriteIndent(indentLevel, "ID: templCSSID,\n"); err != nil {
				return err
			}
			// Class: templ.SafeCSS(".cssID{" + templ.CSSBuilder.String() + "}"),
			if _, err = g.w.WriteIndent(indentLevel, "Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeTemplate(t parser.HTMLTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(t.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Name, r)
	// (
	if _, err = g.w.Write("("); err != nil {
		return err
	}
	// Write parameters.
	if r, err = g.w.Write(t.Parameters.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Parameters, r)
	// ) templ.Component {
	if _, err = g.w.Write(") templ.Component {\n"); err != nil {
		return err
	}
	indentLevel++
	// return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) error {
	if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {\n"); err != nil {
		return err
	}
	{
		indentLevel++
		// ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		if _, err = g.w.WriteIndent(indentLevel, "ctx, _ = templ.RenderedCSSClassesFromContext(ctx)\n"); err != nil {
			return err
		}
		// ctx, _ = templ.RenderedScriptsFromContext(ctx)
		if _, err = g.w.WriteIndent(indentLevel, "ctx, _ = templ.RenderedScriptsFromContext(ctx)\n"); err != nil {
			return err
		}
		// Create StringWriter.
		if _, err = g.w.WriteIndent(indentLevel, "w, ok := writer.(io.StringWriter)\n"); err != nil {
			return err
		}
		if _, err = g.w.WriteIndent(indentLevel, "if !ok {\n"); err != nil {
			return err
		}
		{
			indentLevel++
			if _, err = g.w.WriteIndent(indentLevel, "templw := bufio.NewWriter(writer)\n"); err != nil {
				return err
			}
			if _, err = g.w.WriteIndent(indentLevel, "w = templw\n"); err != nil {
				return err
			}
			if _, err = g.w.WriteIndent(indentLevel, "defer templw.Flush()\n"); err != nil {
				return err
			}
			indentLevel--
		}
		if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
			return err
		}
		// Nodes.
		if err = g.writeNodes(indentLevel, t.Children); err != nil {
			return err
		}
		// return nil
		if _, err = g.w.WriteIndent(indentLevel, "return err\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// })
	if _, err = g.w.WriteIndent(indentLevel, "})\n"); err != nil {
		return err
	}
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeNodes(indentLevel int, nodes []parser.Node) error {
	for i := 0; i < len(nodes); i++ {
		if err := g.writeNode(indentLevel, nodes[i]); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) writeNode(indentLevel int, node parser.Node) error {
	switch n := node.(type) {
	case parser.DocType:
		g.writeDocType(indentLevel, n)
	case parser.Element:
		g.writeElement(indentLevel, n)
	case parser.ForExpression:
		g.writeForExpression(indentLevel, n)
	case parser.CallTemplateExpression:
		g.writeCallTemplateExpression(indentLevel, n)
	case parser.IfExpression:
		g.writeIfExpression(indentLevel, n)
	case parser.SwitchExpression:
		g.writeSwitchExpression(indentLevel, n)
	case parser.StringExpression:
		g.writeStringExpression(indentLevel, n.Expression)
	case parser.Whitespace:
		// Whitespace is not included in template output to minify HTML.
	case parser.Text:
		g.writeText(indentLevel, n)
	default:
		g.w.Write(fmt.Sprintf("Unhandled type: %v\n", reflect.TypeOf(n)))
	}
	return nil
}

func (g *generator) writeDocType(indentLevel int, n parser.DocType) error {
	var err error
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf("_, err = w.WriteString(`<!doctype %s>`)\n", n.Value)); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeIfExpression(indentLevel int, n parser.IfExpression) error {
	var r parser.Range
	var err error
	// if
	if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
		return err
	}
	// x == y
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// Then.
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}
	indentLevel++
	g.writeNodes(indentLevel, n.Then)
	indentLevel--
	if len(n.Else) > 0 {
		// } else {
		if _, err = g.w.WriteIndent(indentLevel, `} else {`+"\n"); err != nil {
			return err
		}
		indentLevel++
		g.writeNodes(indentLevel, n.Else)
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeSwitchExpression(indentLevel int, n parser.SwitchExpression) error {
	var r parser.Range
	var err error
	// switch
	if _, err = g.w.WriteIndent(indentLevel, `switch `); err != nil {
		return err
	}
	// val
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}

	if len(n.Cases) > 0 {
		for _, c := range n.Cases {
			// case
			if _, err = g.w.WriteIndent(indentLevel, `case `); err != nil {
				return err
			}
			// val
			if r, err = g.w.Write(c.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(c.Expression, r)
			if _, err = g.w.Write(`:` + "\n"); err != nil {
				return err
			}
			indentLevel++
			g.writeNodes(indentLevel, c.Children)
			indentLevel--
		}
	}

	if len(n.Default) > 0 {
		if _, err = g.w.WriteIndent(indentLevel, `default:`); err != nil {
			return err
		}
		indentLevel++
		g.writeNodes(indentLevel, n.Default)
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeCallTemplateExpression(indentLevel int, n parser.CallTemplateExpression) error {
	var r parser.Range
	var err error
	if r, err = g.w.WriteIndent(indentLevel, `err = `); err != nil {
		return err
	}
	// Template expression.
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// .Render(ctx, w.(io.Writer))
	if _, err = g.w.Write(".Render(ctx, w.(io.Writer))\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeForExpression(indentLevel int, n parser.ForExpression) error {
	var r parser.Range
	var err error
	// for
	if _, err = g.w.WriteIndent(indentLevel, `for `); err != nil {
		return err
	}
	// i, v := range p.Stuff
	if r, err = g.w.Write(n.Expression.Value); err != nil {
		return err
	}
	g.sourceMap.Add(n.Expression, r)
	// {
	if _, err = g.w.Write(` {` + "\n"); err != nil {
		return err
	}
	// Children.
	indentLevel++
	g.writeNodes(indentLevel, n.Children)
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeErrorHandler(indentLevel int) (err error) {
	_, err = g.w.WriteIndent(indentLevel, "if err != nil {\n")
	if err != nil {
		return err
	}
	indentLevel++
	_, err = g.w.WriteIndent(indentLevel, "return err\n")
	if err != nil {
		return err
	}
	indentLevel--
	_, err = g.w.WriteIndent(indentLevel, "}\n")
	if err != nil {
		return err
	}
	return err
}

func (g *generator) writeElement(indentLevel int, n parser.Element) error {
	if n.IsVoidElement() {
		return g.writeVoidElement(indentLevel, n)
	}
	return g.writeStandardElement(indentLevel, n)
}

func (g *generator) writeVoidElement(indentLevel int, n parser.Element) (err error) {
	if len(n.Children) > 0 {
		return fmt.Errorf("writeVoidElement: void element %q must not have child elements", n.Name)
	}
	if len(n.Attributes) == 0 {
		// <br>
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString("<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		// <hr
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString("<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteIndent(indentLevel, `_, err = w.WriteString(">")`+"\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	}
	return err
}

func (g *generator) writeStandardElement(indentLevel int, n parser.Element) (err error) {
	if len(n.Attributes) == 0 {
		// <div>
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString("<%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	} else {
		// <style type="text/css"></style>
		if err = g.writeElementCSS(indentLevel, n); err != nil {
			return err
		}
		// <script type="text/javascript"></script>
		if err = g.writeElementScript(indentLevel, n); err != nil {
			return err
		}
		// <div
		if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString("<%s")`+"\n", html.EscapeString(n.Name))); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
		if err = g.writeElementAttributes(indentLevel, n); err != nil {
			return err
		}
		// >
		if _, err = g.w.WriteIndent(indentLevel, `_, err = w.WriteString(">")`+"\n"); err != nil {
			return err
		}
		if err = g.writeErrorHandler(indentLevel); err != nil {
			return err
		}
	}
	// Children.
	g.writeNodes(indentLevel, n.Children)
	// </div>
	if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString("</%s>")`+"\n", html.EscapeString(n.Name))); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) writeElementCSS(indentLevel int, n parser.Element) (err error) {
	var r parser.Range
	for i := 0; i < len(n.Attributes); i++ {
		if attr, ok := n.Attributes[i].(parser.ExpressionAttribute); ok {
			name := html.EscapeString(attr.Name)
			if name != "class" {
				continue
			}
			// Create a class name for the style.
			// var templCSSClassess templ.CSSClasses =
			classesName := g.createVariableName()
			if _, err = g.w.WriteIndent(indentLevel, "var "+classesName+" templ.CSSClasses = "); err != nil {
				return err
			}
			// p.Name()
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			if _, err = g.w.Write("\n"); err != nil {
				return err
			}
			// Render the CSS before the element if required.
			// err = templ.RenderCSS(ctx, w, templCSSClassess)
			if _, err = g.w.WriteIndent(indentLevel, "err = templ.RenderCSS(ctx, w, "+classesName+")\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			// Rewrite the ExpressionAttribute to point at the new variable.
			attr.Expression = parser.Expression{
				Value: classesName + ".String()",
			}
			n.Attributes[i] = attr
		}
	}
	return err
}

func (g *generator) writeElementScript(indentLevel int, n parser.Element) (err error) {
	var scriptExpressions []string
	for i := 0; i < len(n.Attributes); i++ {
		if attr, ok := n.Attributes[i].(parser.ExpressionAttribute); ok {
			name := html.EscapeString(attr.Name)
			if strings.HasPrefix(name, "on") {
				scriptExpressions = append(scriptExpressions, attr.Expression.Value)
			}
		}
	}
	// Render the scripts before the element if required.
	// err = templ.RenderScripts(ctx, w, a, b, c)
	if _, err = g.w.WriteIndent(indentLevel, "err = templ.RenderScripts(ctx, w, "+strings.Join(scriptExpressions, ", ")+")\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return err
}

func (g *generator) writeElementAttributes(indentLevel int, n parser.Element) (err error) {
	var r parser.Range
	for i := 0; i < len(n.Attributes); i++ {
		switch attr := n.Attributes[i].(type) {
		case parser.BoolConstantAttribute:
			name := html.EscapeString(attr.Name)
			if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString(" %s")`+"\n", name)); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		case parser.ConstantAttribute:
			name := html.EscapeString(attr.Name)
			value := html.EscapeString(attr.Value)
			if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString(" %s=\"%s\"")`+"\n", name, value)); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		case parser.BoolExpressionAttribute:
			name := html.EscapeString(attr.Name)
			// if
			if _, err = g.w.WriteIndent(indentLevel, `if `); err != nil {
				return err
			}
			// x == y
			if r, err = g.w.Write(attr.Expression.Value); err != nil {
				return err
			}
			g.sourceMap.Add(attr.Expression, r)
			// {
			if _, err = g.w.Write(` {` + "\n"); err != nil {
				return err
			}
			{
				indentLevel++
				if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString(" %s")`+"\n", name)); err != nil {
					return err
				}
				if err = g.writeErrorHandler(indentLevel); err != nil {
					return err
				}
				indentLevel--
			}
			// }
			if _, err = g.w.WriteIndent(indentLevel, `}`+"\n"); err != nil {
				return err
			}
		case parser.ExpressionAttribute:
			name := html.EscapeString(attr.Name)
			// Name
			if _, err = g.w.WriteIndent(indentLevel, fmt.Sprintf(`_, err = w.WriteString(" %s=")`+"\n", name)); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			// Value.
			// Open quote.
			if _, err = g.w.WriteIndent(indentLevel, `_, err = w.WriteString("\"")`+"\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
			if n.Name == "a" && attr.Name == "href" {
				vn := g.createVariableName()
				// var vn templ.SafeURL =
				if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" templ.SafeURL = "); err != nil {
					return err
				}
				// p.Name()
				if r, err = g.w.Write(attr.Expression.Value); err != nil {
					return err
				}
				g.sourceMap.Add(attr.Expression, r)
				if _, err = g.w.Write("\n"); err != nil {
					return err
				}
				if _, err = g.w.WriteIndent(indentLevel, "_, err = w.WriteString(templ.EscapeString(string("+vn+")))\n"); err != nil {
					return err
				}
				if err = g.writeErrorHandler(indentLevel); err != nil {
					return err
				}
			} else {
				if strings.HasPrefix(attr.Name, "on") {
					// It's a JavaScript handler, and requires special handling, because we expect a JavaScript expression.
					vn := g.createVariableName()
					// var vn templ.ComponentScript =
					if _, err = g.w.WriteIndent(indentLevel, "var "+vn+" templ.ComponentScript = "); err != nil {
						return err
					}
					// p.Name()
					if r, err = g.w.Write(attr.Expression.Value); err != nil {
						return err
					}
					g.sourceMap.Add(attr.Expression, r)
					if _, err = g.w.Write("\n"); err != nil {
						return err
					}
					if _, err = g.w.WriteIndent(indentLevel, "_, err = w.WriteString("+vn+".Call)\n"); err != nil {
						return err
					}
					if err = g.writeErrorHandler(indentLevel); err != nil {
						return err
					}
				} else {
					// w.WriteString(templ.EscapeString(
					if _, err = g.w.WriteIndent(indentLevel, "_, err = w.WriteString(templ.EscapeString("); err != nil {
						return err
					}
					// p.Name()
					if r, err = g.w.Write(attr.Expression.Value); err != nil {
						return err
					}
					g.sourceMap.Add(attr.Expression, r)
					// ))
					if _, err = g.w.Write("))\n"); err != nil {
						return err
					}
					if err = g.writeErrorHandler(indentLevel); err != nil {
						return err
					}
				}
			}
			// Close quote.
			if _, err = g.w.WriteIndent(indentLevel, `_, err = w.WriteString("\"")`+"\n"); err != nil {
				return err
			}
			if err = g.writeErrorHandler(indentLevel); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown attribute type %s", reflect.TypeOf(n.Attributes[i]))
		}
	}
	return err
}

func (g *generator) createVariableName() string {
	g.variableID++
	return fmt.Sprintf("var_%d", g.variableID)
}

func (g *generator) writeStringExpression(indentLevel int, e parser.Expression) error {
	var r parser.Range
	var err error
	// w.WriteString(templ.EscapeString(
	if _, err = g.w.WriteIndent(indentLevel, "_, err = w.WriteString(templ.EscapeString("); err != nil {
		return err
	}
	// p.Name()
	if r, err = g.w.Write(e.Value); err != nil {
		return err
	}
	g.sourceMap.Add(e, r)
	// ))
	if _, err = g.w.Write("))\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeText(indentLevel int, e parser.Text) (err error) {
	vn := g.createVariableName()
	// vn := sExpr
	if _, err = g.w.WriteIndent(indentLevel, vn+" := "+createGoString(e.Value)+"\n"); err != nil {
		return err
	}
	// _, err = w.WriteString(vn)
	if _, err = g.w.WriteIndent(indentLevel, "_, err = w.WriteString("+vn+")\n"); err != nil {
		return err
	}
	if err = g.writeErrorHandler(indentLevel); err != nil {
		return err
	}
	return nil
}

func createGoString(s string) string {
	var sb strings.Builder
	sb.WriteRune('`')
	sects := strings.Split(s, "`")
	for i := 0; i < len(sects); i++ {
		sb.WriteString(sects[i])
		if len(sects) > i+1 {
			sb.WriteString("` + \"`\" + `")
		}
	}
	sb.WriteRune('`')
	return sb.String()
}

func (g *generator) writeScript(t parser.ScriptTemplate) error {
	var r parser.Range
	var err error
	var indentLevel int

	// func
	if _, err = g.w.Write("func "); err != nil {
		return err
	}
	if r, err = g.w.Write(t.Name.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Name, r)
	// (
	if _, err = g.w.Write("("); err != nil {
		return err
	}
	// Write parameters.
	if r, err = g.w.Write(t.Parameters.Value); err != nil {
		return err
	}
	g.sourceMap.Add(t.Parameters, r)
	// ) templ.ComponentScript {
	if _, err = g.w.Write(") templ.ComponentScript {\n"); err != nil {
		return err
	}
	indentLevel++
	// return templ.ComponentScript{
	if _, err = g.w.WriteIndent(indentLevel, "return templ.ComponentScript{\n"); err != nil {
		return err
	}
	{
		indentLevel++
		fn := functionName(t.Name.Value, t.Value)
		goFn := createGoString(fn)
		// Name: "scriptName",
		if _, err = g.w.WriteIndent(indentLevel, "Name: "+goFn+",\n"); err != nil {
			return err
		}
		// Function: `function scriptName(a, b, c){` + `constantScriptValue` + `}`,
		prefix := "function " + fn + "(" + stripTypes(t.Parameters.Value) + "){"
		suffix := "}"
		if _, err = g.w.WriteIndent(indentLevel, "Function: "+createGoString(prefix+strings.TrimSpace(t.Value)+suffix)+",\n"); err != nil {
			return err
		}
		// Call: templ.SafeScript(scriptName, a, b, c)
		if _, err = g.w.WriteIndent(indentLevel, "Call: templ.SafeScript("+goFn+", "+stripTypes(t.Parameters.Value)+"),\n"); err != nil {
			return err
		}
		indentLevel--
	}
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n"); err != nil {
		return err
	}
	indentLevel--
	// }
	if _, err = g.w.WriteIndent(indentLevel, "}\n\n"); err != nil {
		return err
	}
	return nil
}

func functionName(name string, body string) string {
	h := sha256.New()
	h.Write([]byte(body))
	hp := hex.EncodeToString(h.Sum(nil))[0:4]
	return fmt.Sprintf("__templ_%s_%s", name, hp)
}

func stripTypes(parameters string) string {
	variableNames := []string{}
	params := strings.Split(parameters, ",")
	for i := 0; i < len(params); i++ {
		p := strings.Split(strings.TrimSpace(params[i]), " ")
		variableNames = append(variableNames, strings.TrimSpace(p[0]))
	}
	return strings.Join(variableNames, ", ")
}
