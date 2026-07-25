package handlers

import "time"

type Base struct {
	SiteName       string
	Categories     []CategoryView
	Authenticated  bool
	ActiveCategory string
	Title          string
	Wide           bool
}

type CategoryView struct {
	ID   int32
	Slug string
	Name string
}

type EntryListItem struct {
	Number       int32
	Title        string
	Slug         string
	Excerpt      string
	CategoryName string
	CategorySlug string
	CreatedAt    time.Time
	IsPrivate    bool
}

type MediaView struct {
	Filepath string
	Filename string
	Kind     string
}

type IngredientView struct {
	Amount string
	Item   string
}

type RecipeView struct {
	Servings    int32
	PrepMinutes int32
	CookMinutes int32
	Ingredients []IngredientView
	Steps       []string
}

type IndexPageData struct {
	Base
	Entries []EntryListItem
}

type CategoryPageData struct {
	Base
	CategorySlug        string
	CategoryName        string
	CategoryDescription string
	Entries             []EntryListItem
}

type EntryPageData struct {
	Base
	Entry       EntryListItem
	ContentHTML string
	Cover       *MediaView
	Gallery     []MediaView
	Recipe      *RecipeView
}

type LoginPageData struct {
	Base
	Error string
}

type AdminEntryRow struct {
	ID           int32
	Number       int32
	Title        string
	Slug         string
	CategoryName string
	TypeLabel    string
	StatusLabel  string
	CreatedAt    time.Time
}

type DashboardPageData struct {
	Base
	Username        string
	Entries         []AdminEntryRow
	AdminCategories []AdminCategoryRow
	CatError        string
}

type VersionSummary struct {
	Number    int32
	Message   string
	CreatedAt time.Time
}

type VersionDetail struct {
	Number      int32
	Message     string
	CreatedAt   time.Time
	Title       string
	Excerpt     string
	ContentHTML string
	EntryType   string
	Recipe      *RecipeView
}

type EntryHistoryPageData struct {
	Base
	EntrySlug  string
	EntryTitle string
	Versions   []VersionSummary
}

type EntryVersionPageData struct {
	Base
	EntrySlug string
	Version   VersionDetail
}

type AdminCategoryRow struct {
	ID          int32
	Slug        string
	Name        string
	Description string
	Accent      string
	EntryCount  int64
}

type AdminEntryVersionsPageData struct {
	Base
	EntryID    int32
	EntryTitle string
	Versions   []VersionSummary
}

type EntryFormPageData struct {
	Base
	Categories         []CategoryView
	IsEdit             bool
	EntryID            int32
	EntryTitle         string
	Excerpt            string
	ContentMarkdown    string
	CategoryID         int32
	EntryType          string
	Published          bool
	IsPrivate          bool
	Recipe             *RecipeView
	ExistingMediaCount int
	CoverFilename      string
}
