package handlers

import (
	"context"
	"database/sql"
	"net/http"

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
