package components

import (
	"fmt"

	"github.com/a-h/templ"
)

func buildUrl(format string, args ...interface{}) string {
	return string(templ.URL(fmt.Sprintf(format, args...)))
}
