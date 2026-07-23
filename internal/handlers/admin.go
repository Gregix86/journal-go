package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sqlc-dev/pqtype"

	"carnet/internal/authx"
	"carnet/internal/db"
	"carnet/internal/mdrender"
	"carnet/internal/slugify"
)

// ---------- auth ----------

func (a *App) LoginForm(w http.ResponseWriter, r *http.Request) {
	data := LoginPageData{Base: a.base(r, "Connexion", "")}
	a.Render.Render(w, "admin/login.html", data)
}

func (a *App) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := a.Queries.GetUserByUsername(r.Context(), username)
	if err != nil || !authx.VerifyPassword(password, user.HashedPassword) {
		data := LoginPageData{Base: a.base(r, "Connexion", ""), Error: "Identifiants incorrects"}
		w.WriteHeader(http.StatusUnauthorized)
		a.Render.Render(w, "admin/login.html", data)
		return
	}

	if err := a.Sessions.Login(w, r, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	_ = a.Sessions.Logout(w, r)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// RequireAdmin protects every /admin/* route except /admin/login.
func (a *App) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Sessions.IsAuthenticated(r) {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ---------- dashboard ----------

func (a *App) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := a.Queries.ListAllEntries(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entries := make([]AdminEntryRow, 0, len(rows))
	for _, row := range rows {
		typeLabel := "Article"
		if row.EntryType == "recipe" {
			typeLabel = "Recette"
		}
		statusLabel := "Brouillon"
		if row.Published {
			statusLabel = "Publie"
		}
		if row.IsPrivate {
			statusLabel += " (prive)"
		}
		entries = append(entries, AdminEntryRow{
			ID: row.ID, Number: row.EntryNumber, Title: row.Title, Slug: row.Slug,
			CategoryName: row.CategoryName, TypeLabel: typeLabel, StatusLabel: statusLabel,
			CreatedAt: row.CreatedAt,
		})
	}

	catRows, err := a.Queries.ListCategories(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cats := make([]AdminCategoryRow, 0, len(catRows))
	for _, c := range catRows {
		count, _ := a.Queries.CountEntriesInCategory(ctx, c.ID)
		cats = append(cats, AdminCategoryRow{
			ID: c.ID, Slug: c.Slug, Name: c.Name, Description: c.Description, Accent: c.Accent, EntryCount: count,
		})
	}

	username := ""
	if uid := a.Sessions.CurrentUserID(r); uid != 0 {
		if u, err := a.userByID(ctx, uid); err == nil {
			username = u
		}
	}

	base := a.base(r, "Tableau de bord", "")
	base.Wide = true
	data := DashboardPageData{
		Base: base, Username: username, Entries: entries,
		AdminCategories: cats, CatError: r.URL.Query().Get("catError"),
	}
	a.Render.Render(w, "admin/dashboard.html", data)
}

func (a *App) userByID(ctx context.Context, id int32) (string, error) {
	// petite requete directe (pas besoin d'une requete sqlc dediee pour ce lookup admin)
	row := a.RawDB.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", id)
	var username string
	err := row.Scan(&username)
	return username, err
}

// ---------- categories ----------

func (a *App) CreateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	if name == "" {
		http.Redirect(w, r, "/admin?catError="+url.QueryEscape("Le nom de la categorie est obligatoire"), http.StatusSeeOther)
		return
	}

	slug := slugify.Slugify(name)
	_, err := a.Queries.CreateCategory(r.Context(), db.CreateCategoryParams{
		Slug: slug, Name: name, Accent: "#1F5C4A", Description: description,
	})
	if err != nil {
		http.Redirect(w, r, "/admin?catError="+url.QueryEscape("Cette categorie existe deja (slug en conflit)"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (a *App) UpdateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	if name == "" {
		http.Redirect(w, r, "/admin?catError="+url.QueryEscape("Le nom de la categorie est obligatoire"), http.StatusSeeOther)
		return
	}

	cat, err := a.Queries.GetCategoryByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Categorie introuvable", http.StatusNotFound)
		return
	}

	err = a.Queries.UpdateCategory(r.Context(), db.UpdateCategoryParams{
		ID: int32(id), Name: name, Description: description, Accent: cat.Accent,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (a *App) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	count, err := a.Queries.CountEntriesInCategory(ctx, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		msg := fmt.Sprintf("Impossible de supprimer : %d entree(s) sont encore rattachee(s) a cette categorie", count)
		http.Redirect(w, r, "/admin?catError="+url.QueryEscape(msg), http.StatusSeeOther)
		return
	}

	if err := a.Queries.DeleteCategory(ctx, int32(id)); err != nil {
		http.Redirect(w, r, "/admin?catError="+url.QueryEscape("Suppression impossible"), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// ---------- entry form ----------

func (a *App) NewEntryForm(w http.ResponseWriter, r *http.Request) {
	cats := a.categories(r.Context())
	data := EntryFormPageData{
		Base: a.base(r, "Nouvelle entree", ""), Categories: cats,
		EntryType: "article", Published: true,
	}
	a.Render.Render(w, "admin/entry_form.html", data)
}

func (a *App) EditEntryForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	row, err := a.Queries.GetEntryByID(ctx, int32(id))
	if err != nil {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}

	var recipe *RecipeView
	mediaCount := 0
	coverName := ""

	if row.EntryType == "recipe" {
		if rd, err := a.Queries.GetRecipeDetail(ctx, row.ID); err == nil {
			var ingredients []IngredientView
			var steps []string
			_ = json.Unmarshal(rd.Ingredients, &ingredients)
			_ = json.Unmarshal(rd.Steps, &steps)
			recipe = &RecipeView{Servings: rd.Servings, PrepMinutes: rd.PrepMinutes, CookMinutes: rd.CookMinutes, Ingredients: ingredients, Steps: steps}
		}
	}

	mediaRows, err := a.Queries.ListMediaByEntry(ctx, sql.NullInt32{Int32: row.ID, Valid: true})
	if err == nil {
		mediaCount = len(mediaRows)
		if row.CoverMediaID.Valid {
			for _, m := range mediaRows {
				if m.ID == row.CoverMediaID.Int32 {
					coverName = m.Filename
				}
			}
		}
	}

	data := EntryFormPageData{
		Base: a.base(r, "Editer l'entree", ""), Categories: a.categories(ctx),
		IsEdit: true, EntryID: row.ID, EntryTitle: row.Title, Excerpt: row.Excerpt,
		ContentMarkdown: row.ContentMarkdown, CategoryID: row.CategoryID, EntryType: row.EntryType,
		Published: row.Published, IsPrivate: row.IsPrivate, Recipe: recipe,
		ExistingMediaCount: mediaCount, CoverFilename: coverName,
	}
	a.Render.Render(w, "admin/entry_form.html", data)
}

const maxUploadBytes = 200 << 20 // 200 MB, adapte selon tes videos

func (a *App) CreateEntry(w http.ResponseWriter, r *http.Request) {
	a.saveEntry(w, r, false)
}

func (a *App) UpdateEntryHandler(w http.ResponseWriter, r *http.Request) {
	a.saveEntry(w, r, true)
}

type recipeFormData struct {
	Servings    int32
	Prep        int32
	Cook        int32
	Ingredients []byte
	Steps       []byte
}

func parseRecipeForm(r *http.Request) recipeFormData {
	amounts := r.Form["ingredient_amount"]
	items := r.Form["ingredient_item"]
	steps := r.Form["step_text"]

	type ing struct {
		Amount string `json:"amount"`
		Item   string `json:"item"`
	}
	var ingredients []ing
	for i := 0; i < len(items); i++ {
		if strings.TrimSpace(items[i]) == "" {
			continue
		}
		amount := ""
		if i < len(amounts) {
			amount = amounts[i]
		}
		ingredients = append(ingredients, ing{Amount: amount, Item: items[i]})
	}
	var cleanSteps []string
	for _, s := range steps {
		if strings.TrimSpace(s) != "" {
			cleanSteps = append(cleanSteps, s)
		}
	}

	ingredientsJSON, _ := json.Marshal(ingredients)
	stepsJSON, _ := json.Marshal(cleanSteps)

	servings, _ := strconv.Atoi(r.FormValue("servings"))
	prep, _ := strconv.Atoi(r.FormValue("prep_minutes"))
	cook, _ := strconv.Atoi(r.FormValue("cook_minutes"))

	return recipeFormData{
		Servings: int32(servings), Prep: int32(prep), Cook: int32(cook),
		Ingredients: ingredientsJSON, Steps: stepsJSON,
	}
}

func (a *App) saveEntry(w http.ResponseWriter, r *http.Request, isEdit bool) {
	ctx := r.Context()

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		http.Error(w, "Formulaire invalide ou fichiers trop volumineux: "+err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	entryType := r.FormValue("entry_type")
	excerpt := r.FormValue("excerpt")
	content := r.FormValue("content_markdown")
	published := r.FormValue("published") != ""
	isPrivate := r.FormValue("is_private") != ""
	versionMessage := strings.TrimSpace(r.FormValue("version_message"))

	var rf recipeFormData
	if entryType == "recipe" {
		rf = parseRecipeForm(r)
	}

	var entryID int32

	if !isEdit {
		baseSlug := slugify.Slugify(title)
		slug := baseSlug
		for i := 2; ; i++ {
			exists, err := a.Queries.EntrySlugExists(ctx, slug)
			if err != nil || !exists {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, i)
		}

		count, err := a.Queries.CountEntries(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		entry, err := a.Queries.CreateEntry(ctx, db.CreateEntryParams{
			CategoryID: int32(categoryID), EntryType: entryType, Title: title, Slug: slug,
			Excerpt: excerpt, ContentMarkdown: content, Published: published, IsPrivate: isPrivate,
			EntryNumber: int32(count) + 1,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		entryID = entry.ID

		msg := versionMessage
		if msg == "" {
			msg = "Version initiale"
		}
		a.createVersionSnapshot(ctx, entryID, 1, title, excerpt, content, entryType, int32(categoryID), rf, msg)
	} else {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		entryID = int32(id)

		old, err := a.Queries.GetEntryByID(ctx, entryID)
		if err != nil {
			http.Error(w, "Entree introuvable", http.StatusNotFound)
			return
		}
		var oldRecipe db.RecipeDetail
		hasOldRecipe := false
		if old.EntryType == "recipe" {
			if rd, err := a.Queries.GetRecipeDetail(ctx, entryID); err == nil {
				oldRecipe = rd
				hasOldRecipe = true
			}
		}

		err = a.Queries.UpdateEntry(ctx, db.UpdateEntryParams{
			ID: entryID, CategoryID: int32(categoryID), EntryType: entryType, Title: title,
			Excerpt: excerpt, ContentMarkdown: content, Published: published, IsPrivate: isPrivate,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		changed := old.Title != title || old.Excerpt != excerpt || old.ContentMarkdown != content ||
			old.EntryType != entryType || old.CategoryID != int32(categoryID)
		if entryType == "recipe" {
			if !hasOldRecipe {
				changed = true
			} else {
				changed = changed || oldRecipe.Servings != rf.Servings || oldRecipe.PrepMinutes != rf.Prep ||
					oldRecipe.CookMinutes != rf.Cook || string(oldRecipe.Ingredients) != string(rf.Ingredients) ||
					string(oldRecipe.Steps) != string(rf.Steps)
			}
		}

		if changed {
			latest, err := a.Queries.GetLatestVersionNumber(ctx, entryID)
			if err != nil {
				latest = 0
			}
			msg := versionMessage
			if msg == "" {
				msg = "Mise a jour"
			}
			a.createVersionSnapshot(ctx, entryID, latest+1, title, excerpt, content, entryType, int32(categoryID), rf, msg)
		}
	}

	if file, header, err := r.FormFile("cover_image"); err == nil {
		defer file.Close()
		media, err := a.saveUpload(ctx, file, header, entryID)
		if err == nil {
			_ = a.Queries.SetEntryCover(ctx, db.SetEntryCoverParams{ID: entryID, CoverMediaID: sql.NullInt32{Int32: media.ID, Valid: true}})
		}
	}

	if r.MultipartForm != nil {
		for _, header := range r.MultipartForm.File["gallery"] {
			file, err := header.Open()
			if err != nil {
				continue
			}
			_, _ = a.saveUpload(ctx, file, header, entryID)
			file.Close()
		}
	}

	if entryType == "recipe" {
		_ = a.Queries.UpsertRecipeDetail(ctx, db.UpsertRecipeDetailParams{
			EntryID: entryID, Servings: rf.Servings, PrepMinutes: rf.Prep, CookMinutes: rf.Cook,
			Ingredients: rf.Ingredients, Steps: rf.Steps,
		})
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func pqtypeRaw(b []byte) pqtype.NullRawMessage {
	if len(b) == 0 {
		return pqtype.NullRawMessage{}
	}
	return pqtype.NullRawMessage{RawMessage: b, Valid: true}
}

func (a *App) createVersionSnapshot(ctx context.Context, entryID, versionNumber int32, title, excerpt, content, entryType string, categoryID int32, rf recipeFormData, message string) {
	params := db.CreateEntryVersionParams{
		EntryID: entryID, VersionNumber: versionNumber, Title: title, Excerpt: excerpt,
		ContentMarkdown: content, EntryType: entryType, CategoryID: categoryID, Message: message,
	}
	if entryType == "recipe" {
		params.RecipeServings = sql.NullInt32{Int32: rf.Servings, Valid: true}
		params.RecipePrepMinutes = sql.NullInt32{Int32: rf.Prep, Valid: true}
		params.RecipeCookMinutes = sql.NullInt32{Int32: rf.Cook, Valid: true}
		params.RecipeIngredients = pqtypeRaw(rf.Ingredients)
		params.RecipeSteps = pqtypeRaw(rf.Steps)
	}
	_, _ = a.Queries.CreateEntryVersion(ctx, params)
}

func (a *App) saveUpload(ctx context.Context, file multipart.File, header *multipart.FileHeader, entryID int32) (db.Medium, error) {
	key, err := authx.GenerateAPIKey() // reutilise comme composant aleatoire pour un nom de fichier unique
	if err != nil {
		return db.Medium{}, err
	}
	ext := filepath.Ext(header.Filename)
	uniqueName := key[:16] + ext
	destPath := filepath.Join(a.Cfg.UploadDir, uniqueName)

	if err := os.MkdirAll(a.Cfg.UploadDir, 0o755); err != nil {
		return db.Medium{}, err
	}
	out, err := os.Create(destPath)
	if err != nil {
		return db.Medium{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		return db.Medium{}, err
	}

	kind := "image"
	contentType := header.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "video") {
		kind = "video"
	}

	return a.Queries.CreateMedia(ctx, db.CreateMediaParams{
		EntryID:  sql.NullInt32{Int32: entryID, Valid: true},
		Filename: header.Filename,
		Filepath: "uploads/" + uniqueName,
		Kind:     kind,
		MimeType: contentType,
	})
}

func (a *App) DeleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	_ = a.Queries.DeleteRecipeDetail(ctx, int32(id))
	_ = a.Queries.DeleteMediaByEntry(ctx, sql.NullInt32{Int32: int32(id), Valid: true})
	_ = a.Queries.DeleteEntry(ctx, int32(id))
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// ---------- version history (admin) ----------

func (a *App) AdminEntryVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	entry, err := a.Queries.GetEntryByID(ctx, int32(id))
	if err != nil {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}

	rows, err := a.Queries.ListEntryVersions(ctx, int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	versions := make([]VersionSummary, 0, len(rows))
	for _, v := range rows {
		versions = append(versions, VersionSummary{Number: v.VersionNumber, Message: v.Message, CreatedAt: v.CreatedAt})
	}

	base := a.base(r, "Historique - "+entry.Title, "")
	base.Wide = true
	data := AdminEntryVersionsPageData{
		Base: base, EntryID: entry.ID, EntryTitle: entry.Title, Versions: versions,
	}
	a.Render.Render(w, "admin/entry_versions.html", data)
}

func (a *App) RestoreEntryVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	versionNumber, _ := strconv.Atoi(chi.URLParam(r, "version"))

	v, err := a.Queries.GetEntryVersion(ctx, db.GetEntryVersionParams{EntryID: int32(id), VersionNumber: int32(versionNumber)})
	if err != nil {
		http.Error(w, "Version introuvable", http.StatusNotFound)
		return
	}

	err = a.Queries.UpdateEntry(ctx, db.UpdateEntryParams{
		ID: int32(id), CategoryID: v.CategoryID, EntryType: v.EntryType, Title: v.Title,
		Excerpt: v.Excerpt, ContentMarkdown: v.ContentMarkdown,
		Published: true, IsPrivate: false, // conserves par defaut ; ajustable ensuite dans le formulaire d'edition
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// on conserve le statut publie/prive courant plutot que de l'ecraser :
	if current, err := a.Queries.GetEntryByID(ctx, int32(id)); err == nil {
		_ = a.Queries.UpdateEntry(ctx, db.UpdateEntryParams{
			ID: int32(id), CategoryID: v.CategoryID, EntryType: v.EntryType, Title: v.Title,
			Excerpt: v.Excerpt, ContentMarkdown: v.ContentMarkdown,
			Published: current.Published, IsPrivate: current.IsPrivate,
		})
	}

	if v.EntryType == "recipe" && v.RecipeIngredients.Valid {
		_ = a.Queries.UpsertRecipeDetail(ctx, db.UpsertRecipeDetailParams{
			EntryID: int32(id), Servings: v.RecipeServings.Int32, PrepMinutes: v.RecipePrepMinutes.Int32,
			CookMinutes: v.RecipeCookMinutes.Int32, Ingredients: v.RecipeIngredients.RawMessage, Steps: v.RecipeSteps.RawMessage,
		})
	}

	latest, err := a.Queries.GetLatestVersionNumber(ctx, int32(id))
	if err != nil {
		latest = int32(versionNumber)
	}
	var rf recipeFormData
	if v.EntryType == "recipe" {
		rf = recipeFormData{
			Servings: v.RecipeServings.Int32, Prep: v.RecipePrepMinutes.Int32, Cook: v.RecipeCookMinutes.Int32,
			Ingredients: v.RecipeIngredients.RawMessage, Steps: v.RecipeSteps.RawMessage,
		}
	}
	msg := fmt.Sprintf("Restauration de la version %d", versionNumber)
	a.createVersionSnapshot(ctx, int32(id), latest+1, v.Title, v.Excerpt, v.ContentMarkdown, v.EntryType, v.CategoryID, rf, msg)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// ---------- version history (public) ----------

func (a *App) EntryHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")

	entry, err := a.Queries.GetEntryBySlug(ctx, slug)
	if err != nil || !entry.Published {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}
	if entry.IsPrivate && !a.Sessions.IsAuthenticated(r) {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}

	rows, err := a.Queries.ListEntryVersions(ctx, entry.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	versions := make([]VersionSummary, 0, len(rows))
	for _, v := range rows {
		versions = append(versions, VersionSummary{Number: v.VersionNumber, Message: v.Message, CreatedAt: v.CreatedAt})
	}

	data := EntryHistoryPageData{
		Base:      a.base(r, "Historique - "+entry.Title, entry.CategorySlug),
		EntrySlug: entry.Slug, EntryTitle: entry.Title, Versions: versions,
	}
	a.Render.Render(w, "entry_history.html", data)
}

func (a *App) EntryVersionPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	versionNumber, _ := strconv.Atoi(chi.URLParam(r, "version"))

	entry, err := a.Queries.GetEntryBySlug(ctx, slug)
	if err != nil || !entry.Published {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}
	if entry.IsPrivate && !a.Sessions.IsAuthenticated(r) {
		http.Error(w, "Entree introuvable", http.StatusNotFound)
		return
	}

	v, err := a.Queries.GetEntryVersion(ctx, db.GetEntryVersionParams{EntryID: entry.ID, VersionNumber: int32(versionNumber)})
	if err != nil {
		http.Error(w, "Version introuvable", http.StatusNotFound)
		return
	}

	contentHTML, err := mdrender.ToHTML(v.ContentMarkdown)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var recipe *RecipeView
	if v.EntryType == "recipe" && v.RecipeIngredients.Valid {
		var ingredients []IngredientView
		var steps []string
		_ = json.Unmarshal(v.RecipeIngredients.RawMessage, &ingredients)
		_ = json.Unmarshal(v.RecipeSteps.RawMessage, &steps)
		recipe = &RecipeView{
			Servings: v.RecipeServings.Int32, PrepMinutes: v.RecipePrepMinutes.Int32, CookMinutes: v.RecipeCookMinutes.Int32,
			Ingredients: ingredients, Steps: steps,
		}
	}

	data := EntryVersionPageData{
		Base:      a.base(r, fmt.Sprintf("Version %d - %s", versionNumber, entry.Title), entry.CategorySlug),
		EntrySlug: entry.Slug,
		Version: VersionDetail{
			Number: v.VersionNumber, Message: v.Message, CreatedAt: v.CreatedAt,
			Title: v.Title, Excerpt: v.Excerpt, ContentHTML: contentHTML, EntryType: v.EntryType, Recipe: recipe,
		},
	}
	a.Render.Render(w, "entry_version.html", data)
}
