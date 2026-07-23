package render

import (
	"fmt"
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"pad3": func(n int32) string {
		return fmt.Sprintf("%03d", n)
	},
	"dateFR": func(t time.Time) string {
		return t.Format("02.01.2006")
	},
	"datetimeFR": func(t time.Time) string {
		return t.Format("02.01.2006 a 15:04")
	},
	"safeHTML": func(s string) template.HTML {
		return template.HTML(s)
	},
	"add": func(a, b int) int { return a + b },
}
