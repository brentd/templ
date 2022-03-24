// Code generated by templ@(devel) DO NOT EDIT.

package testfor

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bufio"

func render(items []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		w, ok := writer.(io.StringWriter)
		if !ok {
			templw := bufio.NewWriter(writer)
			w = templw
			defer templw.Flush()
		}
		for _, item := range items {
			_, err = w.WriteString("<div>")
			if err != nil {
				return err
			}
			_, err = w.WriteString(templ.EscapeString(item))
			if err != nil {
				return err
			}
			_, err = w.WriteString("</div>")
			if err != nil {
				return err
			}
		}
		return err
	})
}

