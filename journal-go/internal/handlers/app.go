package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"carnet/internal/authx"
	"carnet/internal/config"
	"carnet/internal/db"
	"carnet/internal/render"
)

type App struct {
	Queries  *db.Queries
	RawDB    *sql.DB
	Cfg      config.Config
	Sessions *authx.Sessions
	Render   *render.Renderer
}

func (a *App) categories(ctx context.Context) []CategoryView {
	cats, err := a.Queries.ListCategories(ctx)
	if err != nil {
		return nil
	}
	out := make([]CategoryView, 0, len(cats))
	for _, c := range cats {
		out = append(out, CategoryView{ID: c.ID, Slug: c.Slug, Name: c.Name})
	}
	return out
}

func (a *App) base(r *http.Request, title, activeCategory string) Base {
	return Base{
		SiteName:       a.Cfg.SiteName,
		Categories:     a.categories(r.Context()),
		Authenticated:  a.Sessions.IsAuthenticated(r),
		ActiveCategory: activeCategory,
		Title:          title,
	}
}

// parseInt32 convertit une chaine en int32 en verifiant les bornes via
// strconv.ParseInt(s, 10, 32), qui renvoie une erreur si la valeur ne rentre
// pas dans un int32 - contrairement a strconv.Atoi + conversion, qui
// depasserait silencieusement (integer overflow) sur une tres grande valeur.
func parseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32) // #nosec G115 -- bitSize=32 garantit que v tient dans un int32
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}
