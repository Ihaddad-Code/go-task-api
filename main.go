package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Store en mémoire depuis un fichier
	store := NewTaskStore("tasks.json")

	// /tasks (GET, POST)
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListTasks(w, r, store)
		case http.MethodPost:
			handleCreateTask(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// /tasks/{id} (GET, PUT, DELETE)
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			handleGetTask(w, r, store, id)
		case http.MethodPut:
			handleUpdateTask(w, r, store, id)
		case http.MethodDelete:
			handleDeleteTask(w, r, store, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           logRequest(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Serveur HTTP sur %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur serveur: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Serveur arrêté proprement")
}

// --- Handlers JSON ---

type createTaskInput struct {
	Title string `json:"title"`
}

func handleCreateTask(w http.ResponseWriter, r *http.Request, s *TaskStore) {
	defer r.Body.Close()
	var in createTaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Title) == "" {
		http.Error(w, "invalid body (need non-empty title)", http.StatusBadRequest)
		return
	}
	t := s.Create(strings.TrimSpace(in.Title))
	if err := s.Save(); err != nil { 
		http.Error(w, "failed to persist", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func handleListTasks(w http.ResponseWriter, r *http.Request, s *TaskStore) {
	writeJSON(w, http.StatusOK, s.List())
}

func handleGetTask(w http.ResponseWriter, r *http.Request, s *TaskStore, id int64) {
	t, err := s.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

type updateTaskInput struct {
	Title *string `json:"title,omitempty"`
	Done  *bool   `json:"done,omitempty"`
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request, s *TaskStore, id int64) {
	defer r.Body.Close()
	existing, err := s.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var in updateTaskInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			http.Error(w, "title cannot be empty", http.StatusBadRequest)
			return
		}
		existing.Title = title
	}
	if in.Done != nil {
		existing.Done = *in.Done
	}
	if err := s.Update(existing); err != nil {
		http.NotFound(w, r)
		return
	}
	if err := s.Save(); err != nil {
		http.Error(w, "failed to persist", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request, s *TaskStore, id int64) {
	if err := s.Delete(id); err != nil {
		http.NotFound(w, r)
		return
	}
	if err := s.Save(); err != nil {
		http.Error(w, "failed to persist", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}