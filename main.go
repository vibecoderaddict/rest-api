package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type itemStore struct {
	sync.RWMutex
	nextID int
	items  map[int]Item
}

func newHandler() http.Handler {
	store := &itemStore{nextID: 1, items: make(map[int]Item)}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
			return
		}

		if r.URL.Path == "/items" {
			switch r.Method {
			case http.MethodGet:
				store.list(w)
			case http.MethodPost:
				store.create(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasPrefix(r.URL.Path, "/items/") {
			id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/items/"))
			if err != nil {
				writeError(w, http.StatusNotFound, "item not found")
				return
			}
			switch r.Method {
			case http.MethodGet:
				store.get(w, id)
			case http.MethodPut:
				store.update(w, r, id)
			case http.MethodDelete:
				store.delete(w, id)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		writeError(w, http.StatusNotFound, "route not found")
	})
}

func (s *itemStore) list(w http.ResponseWriter) {
	s.RLock()
	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	s.RUnlock()
	writeJSON(w, http.StatusOK, items)
}

func (s *itemStore) get(w http.ResponseWriter, id int) {
	s.RLock()
	item, ok := s.items[id]
	s.RUnlock()
	if !ok {
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *itemStore) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil || strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.Lock()
	item := Item{ID: s.nextID, Name: strings.TrimSpace(input.Name)}
	s.nextID++
	s.items[item.ID] = item
	s.Unlock()
	writeJSON(w, http.StatusCreated, item)
}

func (s *itemStore) update(w http.ResponseWriter, r *http.Request, id int) {
	var input struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &input); err != nil || strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	s.Lock()
	if _, ok := s.items[id]; !ok {
		s.Unlock()
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	item := Item{ID: id, Name: strings.TrimSpace(input.Name)}
	s.items[id] = item
	s.Unlock()
	writeJSON(w, http.StatusOK, item)
}

func (s *itemStore) delete(w http.ResponseWriter, id int) {
	s.Lock()
	if _, ok := s.items[id]; !ok {
		s.Unlock()
		writeError(w, http.StatusNotFound, "item not found")
		return
	}
	delete(s.items, id)
	s.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) == nil {
		return errors.New("multiple JSON values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
