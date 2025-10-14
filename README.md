# go-task-api

Mini API REST en **Go** pour gérer des tâches (Todo).  
✅ Concurrence thread-safe (`RWMutex`)  
✅ Persistance simple via `tasks.json`  
✅ Tests unitaires (`testing`, `httptest`)  
✅ Docker multi-stage (image légère)

## 📦 Endpoints

| Méthode | URL           | Description               | Body (JSON)                      |
| ------: | ------------- | ------------------------- | -------------------------------- |
|     GET | `/healthz`    | Statut du service         | —                                |
|     GET | `/tasks`      | Liste toutes les tâches   | —                                |
|    POST | `/tasks`      | Crée une tâche            | `{ "title": "Texte..." }`        |
|     GET | `/tasks/{id}` | Récupère une tâche par ID | —                                |
|     PUT | `/tasks/{id}` | Met à jour (title/done)   | `{ "title":"...", "done":true }` |
|  DELETE | `/tasks/{id}` | Supprime une tâche        | —                                |

## 🚀 Lancer en local

```bash
go run .
# ping
curl http://localhost:8080/healthz
```
