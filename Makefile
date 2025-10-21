# Nom de l'image Docker
IMAGE_NAME = go-task-api

# Exécuter le projet en local
run: 
	go run .

# Construire un binaire local
build:
	go build -o go-task-api .

# --- Docker ---

# Construire l'image Docker
docker-build:
	docker buildx build --platform linux/arm64 -t $(IMAGE_NAME) .

# Lancer le conteneur (sans persistance)
docker-run:
	docker run --rm -p 8080:8080 $(IMAGE_NAME)

# Lancer le conteneur avec persistance via tasks.json
docker-run-persist:
	@touch tasks.json
	docker run --rm -p 8080:8080 \
		-v "$(PWD)/tasks.json:/app/tasks.json" \
		$(IMAGE_NAME)

# Supprimer l'image Docker
docker-clean:
	docker rmi $(IMAGE_NAME) || true

# Nettoyer le binaire local
clean:
	rm -f go-task-api