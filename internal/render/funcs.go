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
		return template.HTML(s) // #nosec G203 -- contenu genere par notre rendu Markdown (goldmark) a partir d'articles ecrits par les administrateurs authentifies, jamais par un visiteur anonyme
	},
	"add": func(a, b int) int { return a + b },
}
