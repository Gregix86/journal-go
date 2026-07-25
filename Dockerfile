# ---- etape de build ----
FROM golang:1.25-bookworm AS build
 
WORKDIR /src
 
COPY go.mod go.sum ./
RUN go mod download
 
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/carnet ./cmd/server
 
# ---- image finale ----
# Variante "nonroot" : le conteneur tourne avec un utilisateur non-root
# (uid/gid 65532) plutot que root, par securite.
FROM gcr.io/distroless/static-debian12:nonroot
 
WORKDIR /app
COPY --from=build --chown=nonroot:nonroot /out/carnet ./carnet
COPY --chown=nonroot:nonroot templates ./templates
COPY --chown=nonroot:nonroot static ./static
 
EXPOSE 8000
USER nonroot:nonroot
ENTRYPOINT ["./carnet"]