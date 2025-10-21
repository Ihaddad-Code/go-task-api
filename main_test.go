package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func newTestMux(store *TaskStore) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListTasks(w, store)
		case http.MethodPost:
			handleCreateTask(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
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
	return mux
}

func TestCreateInvalidBody(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvalidID(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/9999", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateAndListTasks(t *testing.T) {
	store := NewTaskStore("")
	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleListTasks(w, store)
		case http.MethodPost:
			handleCreateTask(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	body := bytes.NewBufferString(`{"title":"Apprendre Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	var created Task
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if created.Title != "Apprendre Go" {
		t.Errorf("expected title 'Apprendre Go', got '%s'", created.Title)
	}

	req = httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var list []Task
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 task, got %d", len(list))
	}
}

func TestUpdateTask(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	body := bytes.NewBufferString(`{"title":"Apprendre Go"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d", rec.Code)
	}
	var created Task
	_ = json.NewDecoder(rec.Body).Decode(&created)

	updateBody := bytes.NewBufferString(`{"title":"Apprendre Go (v2)", "done":true}`)
	req = httptest.NewRequest(http.MethodPut, "/tasks/"+strconv.FormatInt(created.ID, 10), updateBody)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d", rec.Code)
	}
	var updated Task
	_ = json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Title != "Apprendre Go (v2)" || updated.Done != true {
		t.Fatalf("unexpected update: %+v", updated)
	}
}

func TestDeleteTask(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	body := bytes.NewBufferString(`{"title":"A supprimer"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create expected 201, got %d", rec.Code)
	}
	var created Task
	_ = json.NewDecoder(rec.Body).Decode(&created)

	// delete
	req = httptest.NewRequest(http.MethodDelete, "/tasks/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete expected 204, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/tasks/"+strconv.FormatInt(created.ID, 10), nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get after delete expected 404, got %d", rec.Code)
	}
}

func TestPersistenceSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.json")
	
	storeA := NewTaskStore(path)
	storeA.Create("Persistante")
	if err := storeA.Save(); err != nil {
		t.Fatalf("save error: %v", err)
	}

	storeB := NewTaskStore(path)
	list := storeB.List()
	if len(list) != 1 || list[0].Title != "Persistante" {
		t.Fatalf("expected 1 task 'Persistante', got %+v", list)
	}
}

func TestUpdateEmptyTitle(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	body := bytes.NewBufferString(`{"title":"Init"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	var created Task
	_ = json.NewDecoder(rec.Body).Decode(&created)

	updateBody := bytes.NewBufferString(`{"title":"   "}`)
	req = httptest.NewRequest(http.MethodPut, "/tasks/"+strconv.FormatInt(created.ID, 10), updateBody)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty title, got %d", rec.Code)
	}
}

func TestUpdateNotFound(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	updateBody := bytes.NewBufferString(`{"title":"Ghost"}`)
	req := httptest.NewRequest(http.MethodPut, "/tasks/9999", updateBody)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteNotFound(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1234", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestLogRequestMiddleware(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)
	handler := logRequest(mux)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCreateMalformedJSON(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{`)) // JSON cassé
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateMalformedJSON(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	body := bytes.NewBufferString(`{"title":"Init"}`)
	req := httptest.NewRequest(http.MethodPost, "/tasks", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var created Task
	_ = json.NewDecoder(rec.Body).Decode(&created)

	req = httptest.NewRequest(http.MethodPut, "/tasks/"+strconv.FormatInt(created.ID, 10), bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestInvalidIDZeroNegative(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	for _, bad := range []string{"0", "-1"} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/tasks/"+bad, nil)
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("id %s expected 400, got %d", bad, rec.Code)
		}
	}
}

func TestMethodNotAllowed(t *testing.T) {
	store := NewTaskStore("")
	mux := newTestMux(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/tasks", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/tasks/1", nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestPersistenceErrorsReturn500(t *testing.T) {
	dir := t.TempDir()
	store := NewTaskStore(dir)
	mux := newTestMux(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBufferString(`{"title":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("POST expected 500, got %d", rec.Code)
	}

	task := store.Create("temp")
	
	body := bytes.NewBufferString(`{"done":true}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/tasks/"+strconv.FormatInt(task.ID, 10), body)
	req.Header.Set("Content-Type", "application/json")
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("PUT expected 500, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/tasks/"+strconv.FormatInt(task.ID, 10), nil)
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("DELETE expected 500, got %d", rec.Code)
	}
}