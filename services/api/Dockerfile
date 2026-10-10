FROM golang:1.27.2-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS builder
WORKDIR /app

# Cache dependency download layer separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Distroless: no shell, minimal attack surface, non-root by default
FROM gcr.io/distroless/static-debian12@sha256:20bc6c0bc4d625a22a8fde3e55f6515709b32055ef8fb9cfbddaa06d1760f838
COPY --from=builder /app/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
