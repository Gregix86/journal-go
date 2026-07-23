# ---- etape de build ----
FROM golang:1.22-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/carnet ./cmd/server

# ---- image finale ----
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/carnet ./carnet
COPY templates ./templates
COPY static ./static

EXPOSE 8000
ENTRYPOINT ["./carnet"]
