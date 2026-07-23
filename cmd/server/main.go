package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"

	"carnet/internal/authx"
	"carnet/internal/config"
	"carnet/internal/db"
	"carnet/internal/dbinit"
	"carnet/internal/handlers"
	"carnet/internal/render"
)

func main() {
	cfg := config.Load()

	rawDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connexion base de donnees: %v", err)
	}
	defer rawDB.Close()

	if err := rawDB.Ping(); err != nil {
		log.Fatalf("ping base de donnees: %v", err)
	}

	if err := dbinit.Apply(rawDB); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	queries := db.New(rawDB)

	if err := seed(rawDB, queries, cfg); err != nil {
		log.Fatalf("seed initial: %v", err)
	}

	renderer, err := render.New("templates")
	if err != nil {
		log.Fatalf("chargement des templates: %v", err)
	}

	app := &handlers.App{
		Queries:  queries,
		RawDB:    rawDB,
		Cfg:      cfg,
		Sessions: authx.NewSessions(cfg.SecretKey),
		Render:   renderer,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fileServer := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// ---- public ----
	r.Get("/", app.Home)
	r.Get("/categorie/{slug}", app.CategoryPage)
	r.Get("/entree/{slug}", app.EntryPage)
	r.Get("/entree/{slug}/historique", app.EntryHistory)
	r.Get("/entree/{slug}/historique/{version}", app.EntryVersionPage)

	// ---- admin ----
	r.Get("/admin/login", app.LoginForm)
	r.Post("/admin/login", app.LoginSubmit)
	r.Get("/admin/logout", app.Logout)

	r.Route("/admin", func(r chi.Router) {
		r.Use(app.RequireAdmin)
		r.Get("/", app.Dashboard)
		r.Get("/entries/new", app.NewEntryForm)
		r.Post("/entries/new", app.CreateEntry)
		r.Get("/entries/{id}/edit", app.EditEntryForm)
		r.Post("/entries/{id}/edit", app.UpdateEntryHandler)
		r.Post("/entries/{id}/delete", app.DeleteEntryHandler)
		r.Get("/entries/{id}/versions", app.AdminEntryVersions)
		r.Post("/entries/{id}/versions/{version}/restore", app.RestoreEntryVersion)
		r.Post("/categories/new", app.CreateCategoryHandler)
		r.Post("/categories/{id}/edit", app.UpdateCategoryHandler)
		r.Post("/categories/{id}/delete", app.DeleteCategoryHandler)
	})

	addr := ":" + cfg.Port
	log.Printf("Le carnet ecoute sur %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func seed(rawDB *sql.DB, queries *db.Queries, cfg config.Config) error {
	ctx := context.Background()

	type catSeed struct{ slug, name, accent, desc string }
	defaults := []catSeed{
		{"technologie", "Technologie", "#1F5C4A", "Projets et bidouilles"},
		{"cuisine", "Cuisine", "#1F5C4A", "Recettes testees et approuvees"},
	}
	for _, c := range defaults {
		if err := queries.CreateCategoryIfMissing(ctx, db.CreateCategoryIfMissingParams{
			Slug: c.slug, Name: c.name, Accent: c.accent, Description: c.desc,
		}); err != nil {
			return err
		}
	}

	if hash, err := authx.HashPassword(cfg.AdminPassword); err == nil {
		_, _ = queries.CreateUser(ctx, db.CreateUserParams{Username: cfg.AdminUsername, HashedPassword: hash})
	}
	if cfg.Admin2Username != "" {
		if hash, err := authx.HashPassword(cfg.Admin2Password); err == nil {
			_, _ = queries.CreateUser(ctx, db.CreateUserParams{Username: cfg.Admin2Username, HashedPassword: hash})
		}
	}

	return nil
}
