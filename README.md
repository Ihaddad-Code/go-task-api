# 🧩 go-task-api

Mini API REST en **Go** pour gérer des tâches (Todo).  
✅ Concurrence thread-safe (`RWMutex`)  
✅ Persistance simple via `tasks.json`  
✅ Tests unitaires (`testing`, `httptest`)  
✅ Docker multi-stage (image légère)  
✅ Makefile pour un lancement rapide

---

## 📦 Endpoints

| Méthode    | URL           | Description               | Body (JSON)                      |
| ---------- | ------------- | ------------------------- | -------------------------------- |
| **GET**    | `/healthz`    | Statut du service         | —                                |
| **GET**    | `/tasks`      | Liste toutes les tâches   | —                                |
| **POST**   | `/tasks`      | Crée une tâche            | `{ "title": "Texte..." }`        |
| **GET**    | `/tasks/{id}` | Récupère une tâche par ID | —                                |
| **PUT**    | `/tasks/{id}` | Met à jour (title/done)   | `{ "title":"...", "done":true }` |
| **DELETE** | `/tasks/{id}` | Supprime une tâche        | —                                |

---

## 🚀 Lancer en local

```bash
go run .
# Vérifier le serveur
curl http://localhost:8080/healthz
```

---

## 1 Lancer via le Makefile avec et sans Docker

### Avec Docker

```bash
make docker-build
make docker-run-persist
```

### Sans Docker

```bash
make run
```

## 2 Dans un nouveau terminal

```bash
# 1) Vérifier le service
curl -s http://localhost:8080/healthz

# 2) Créer 3 tâches
curl -s -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Implémentation du server"}'
curl -s -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Écrire des tests"}'
curl -s -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Dockeriser le projet"}'

# 3) Lister toutes les tâches
curl -s http://localhost:8080/tasks

# 4) Marquer la tâche #1 comme terminée (done=true)
curl -s -X PUT http://localhost:8080/tasks/1 -H "Content-Type: application/json" -d '{"done":true}'

# 5) Mettre à jour le titre de la tâche #2 (et la laisser non terminée)
curl -s -X PUT http://localhost:8080/tasks/2 -H "Content-Type: application/json" -d '{"title":"Écrire des tests (unitaires)"}'

# 6) Vérifier l’état actuel
curl -s http://localhost:8080/tasks

# 7) Supprimer une tâche (ex: #3)
curl -s -X DELETE http://localhost:8080/tasks/3 -i

# 8) Supprimer plusieurs tâches (ex: #1 et #2)
for id in 1 2; do
  curl -s -X DELETE http://localhost:8080/tasks/$id -i
done

# 9) Confirmer qu’il n’y a plus de tâches
curl -s http://localhost:8080/tasks
```
