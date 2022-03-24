// Code generated by templ@(devel) DO NOT EDIT.

package testcall

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bufio"

func personTemplate(p person) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		w, ok := writer.(io.StringWriter)
		if !ok {
			templw := bufio.NewWriter(writer)
			w = templw
			defer templw.Flush()
		}
		_, err = w.WriteString("<div>")
		if err != nil {
			return err
		}
		_, err = w.WriteString("<h1>")
		if err != nil {
			return err
		}
		_, err = w.WriteString(templ.EscapeString(p.name))
		if err != nil {
			return err
		}
		_, err = w.WriteString("</h1>")
		if err != nil {
			return err
		}
		err = templ.RenderScripts(ctx, w, )
		if err != nil {
			return err
		}
		_, err = w.WriteString("<div")
		if err != nil {
			return err
		}
		_, err = w.WriteString(" style=\"font-family: &#39;sans-serif&#39;\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(" id=\"test\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(" data-contents=")
		if err != nil {
			return err
		}
		_, err = w.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(templ.EscapeString(`something with "quotes" and a <tag>`))
		if err != nil {
			return err
		}
		_, err = w.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(">")
		if err != nil {
			return err
		}
		err = email(p.email).Render(ctx, w.(io.Writer))
		if err != nil {
			return err
		}
		_, err = w.WriteString("</div>")
		if err != nil {
			return err
		}
		_, err = w.WriteString("</div>")
		if err != nil {
			return err
		}
		return err
	})
}

func email(s string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		w, ok := writer.(io.StringWriter)
		if !ok {
			templw := bufio.NewWriter(writer)
			w = templw
			defer templw.Flush()
		}
		_, err = w.WriteString("<div>")
		if err != nil {
			return err
		}
		var_1 := `email:`
		_, err = w.WriteString(var_1)
		if err != nil {
			return err
		}
		err = templ.RenderScripts(ctx, w, )
		if err != nil {
			return err
		}
		_, err = w.WriteString("<a")
		if err != nil {
			return err
		}
		_, err = w.WriteString(" href=")
		if err != nil {
			return err
		}
		_, err = w.WriteString("\"")
		if err != nil {
			return err
		}
		var var_2 templ.SafeURL = templ.URL("mailto: " + s)
		_, err = w.WriteString(templ.EscapeString(string(var_2)))
		if err != nil {
			return err
		}
		_, err = w.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(">")
		if err != nil {
			return err
		}
		_, err = w.WriteString(templ.EscapeString(s))
		if err != nil {
			return err
		}
		_, err = w.WriteString("</a>")
		if err != nil {
			return err
		}
		_, err = w.WriteString("</div>")
		if err != nil {
			return err
		}
		return err
	})
}

