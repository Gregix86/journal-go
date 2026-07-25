package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

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
	r.Use(securityHeaders)
	// Protection CSRF native (Go 1.25+) : bloque les requetes cross-origin qui
	// changent l'etat (POST/PUT/DELETE...) en verifiant Sec-Fetch-Site/Origin.
	// Pas de jeton/cookie a gerer, contrairement aux anciennes solutions.
	r.Use(http.NewCrossOriginProtection().Handler)

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
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
		// ReadHeaderTimeout protege contre les attaques "slowloris" (connexions
		// ouvertes qui envoient les en-tetes tres lentement). Pas de ReadTimeout
		// global car les uploads de photos/videos peuvent prendre du temps.
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("Le carnet ecoute sur %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src https://fonts.gstatic.com; "+
				"img-src 'self' data:; "+
				"script-src 'self' 'unsafe-inline'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'; "+
				"object-src 'none'")
		next.ServeHTTP(w, r)
	})
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