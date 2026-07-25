package config

import "os"

type Config struct {
	DatabaseURL    string
	SecretKey      string
	SiteName       string
	UploadDir      string
	AdminUsername  string
	AdminPassword  string
	Admin2Username string
	Admin2Password string
	Port           string
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func Load() Config {
	return Config{
		DatabaseURL:    envOr("DATABASE_URL", "postgres://journal:journal@db:5432/journal?sslmode=disable"),
		SecretKey:      envOr("SECRET_KEY", "change-me-in-production"),
		SiteName:       envOr("SITE_NAME", "Mon Carnet"),
		UploadDir:      envOr("UPLOAD_DIR", "static/uploads"),
		AdminUsername:  envOr("ADMIN_USERNAME", "admin"),
		AdminPassword:  envOr("ADMIN_PASSWORD", "changeme"),
		Admin2Username: envOr("ADMIN2_USERNAME", ""),
		Admin2Password: envOr("ADMIN2_PASSWORD", ""),
		Port:           envOr("PORT", "8000"),
	}
}
