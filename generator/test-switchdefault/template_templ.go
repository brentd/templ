// Code generated by templ@(devel) DO NOT EDIT.

package testswitchdefault

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bufio"

func template(input string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, writer io.Writer) (err error) {
		ctx, _ = templ.RenderedCSSClassesFromContext(ctx)
		ctx, _ = templ.RenderedScriptsFromContext(ctx)
		w, ok := writer.(io.StringWriter)
		if !ok {
			templw := bufio.NewWriter(writer)
			w = templw
			defer templw.Flush()
		}
		switch input {
		case "a":
			_, err = w.WriteString(templ.EscapeString("it was 'a'"))
			if err != nil {
				return err
			}
		default:			_, err = w.WriteString(templ.EscapeString("it was something else"))
			if err != nil {
				return err
			}
		}
		return err
	})
}

