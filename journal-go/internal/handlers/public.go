package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"carnet/internal/mdrender"
)

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	authed := a.Sessions.IsAuthenticated(r)

	var items []EntryListItem
	if authed {
		rows, err := a.Queries.ListLatestAnyVisibilityEntries(ctx, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			items = append(items, EntryListItem{
				Number: row.EntryNumber, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
				CategoryName: row.CategoryName, CategorySlug: row.CategorySlug,
				CreatedAt: row.CreatedAt, IsPrivate: row.IsPrivate,
			})
		}
	} else {
		rows, err := a.Queries.ListLatestPublicEntries(ctx, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			items = append(items, EntryListItem{
				Number: row.EntryNumber, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
				CategoryName: row.CategoryName, CategorySlug: row.CategorySlug,
				CreatedAt: row.CreatedAt, IsPrivate: row.IsPrivate,
			})
		}
	}

	data := IndexPageData{Base: a.base(r, "", ""), Entries: items}
	a.Render.Render(w, "index.html", data)
}

func (a *App) CategoryPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	cat, err := a.Queries.GetCategoryBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "Categorie introuvable", http.StatusNotFound)
		return
	}

	authed := a.Sessions.IsAuthenticated(r)
	var items []EntryListItem
	if authed {
		rows, err := a.Queries.ListEntriesByCategoryAnyVisibility(ctx, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			items = append(items, EntryListItem{
				Number: row.EntryNumber, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
				CreatedAt: row.CreatedAt, IsPrivate: row.IsPrivate,
			})
		}
	} else {
		rows, err := a.Queries.ListEntriesByCategoryPublic(ctx, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, row := range rows {
			items = append(items, EntryListItem{
				Number: row.EntryNumber, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
				CreatedAt: row.CreatedAt, IsPrivate: row.IsPrivate,
			})
		}
	}

	data := CategoryPageData{
		Base:                a.base(r, cat.Name, cat.Slug),
		CategorySlug:        cat.Slug,
		CategoryName:        cat.Name,
		CategoryDescription: cat.Description,
		Entries:             items,
	}
	a.Render.Render(w, "category.html", data)
}

func (a *App) EntryPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	row, err := a.Queries.GetEntryBySlug(ctx, slug)
	if err != nil {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}
	if !row.Published {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}
	if row.IsPrivate && !a.Sessions.IsAuthenticated(r) {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}

	contentHTML, err := mdrender.ToHTML(row.ContentMarkdown)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mediaRows, err := a.Queries.ListMediaByEntry(ctx, sql.NullInt32{Int32: row.ID, Valid: true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cover *MediaView
	var gallery []MediaView
	for _, m := range mediaRows {
		mv := MediaView{Filepath: m.Filepath, Filename: m.Filename, Kind: m.Kind}
		if row.CoverMediaID.Valid && m.ID == row.CoverMediaID.Int32 {
			c := mv
			cover = &c
			continue
		}
		gallery = append(gallery, mv)
	}

	var recipe *RecipeView
	if row.EntryType == "recipe" {
		rd, err := a.Queries.GetRecipeDetail(ctx, row.ID)
		if err == nil {
			var ingredients []IngredientView
			var steps []string
			_ = json.Unmarshal(rd.Ingredients, &ingredients)
			_ = json.Unmarshal(rd.Steps, &steps)
			recipe = &RecipeView{
				Servings: rd.Servings, PrepMinutes: rd.PrepMinutes, CookMinutes: rd.CookMinutes,
				Ingredients: ingredients, Steps: steps,
			}
		}
	}

	entry := EntryListItem{
		Number: row.EntryNumber, Title: row.Title, Slug: row.Slug, Excerpt: row.Excerpt,
		CategoryName: row.CategoryName, CategorySlug: row.CategorySlug,
		CreatedAt: row.CreatedAt, IsPrivate: row.IsPrivate,
	}

	data := EntryPageData{
		Base:        a.base(r, row.Title, row.CategorySlug),
		Entry:       entry,
		ContentHTML: contentHTML,
		Cover:       cover,
		Gallery:     gallery,
		Recipe:      recipe,
	}
	a.Render.Render(w, "entry.html", data)
}
