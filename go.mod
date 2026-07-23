module carnet

go 1.22.2

replace golang.org/x/crypto => github.com/golang/crypto v0.31.0

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/gorilla/sessions v1.3.0
	github.com/lib/pq v1.12.3
	github.com/yuin/goldmark v1.8.4
	golang.org/x/crypto v0.31.0
)

require (
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/sqlc-dev/pqtype v0.3.0 // indirect
)
