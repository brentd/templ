// Code generated by templ@(devel) DO NOT EDIT.

package testvoid

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bufio"

func render() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		w, ok := writer.(io.StringWriter)
		if !ok {
			templw := bufio.NewWriter(writer)
			w = templw
			defer templw.Flush()
		}
		_, err = w.WriteString("<br>")
		if err != nil {
			return err
		}
		_, err = w.WriteString("<img")
		if err != nil {
			return err
		}
		_, err = w.WriteString(" src=\"https://example.com/image.png\"")
		if err != nil {
			return err
		}
		_, err = w.WriteString(">")
		if err != nil {
			return err
		}
		_, err = w.WriteString("<br>")
		if err != nil {
			return err
		}
		_, err = w.WriteString("<br>")
		if err != nil {
			return err
		}
		return err
	})
}

