# builder (compilation)
FROM golang:1.25.2-alpine AS builder

ENV GOTOOLCHAIN=auto

# Réduit la taille de l’image
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copie les fichiers du projet
COPY go.mod ./
RUN go mod download

COPY . .

# Compile le binaire statiquement (aucune dépendance)
RUN CGO_ENABLED=0 GOOS=linux go build -o go-task-api .

# image finale minimale
FROM alpine:latest

# Ajoute le certificat SSL pour curl, etc.
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copie uniquement le binaire
COPY --from=builder /app/go-task-api .

EXPOSE 8080

# Commande de démarrage
CMD ["./go-task-api"]