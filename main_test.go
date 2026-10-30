package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestItemsCRUD(t *testing.T) {
	handler := newHandler()

	create := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"learn Go"}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", created.Code, http.StatusCreated)
	}

	var item Item
	if err := json.NewDecoder(created.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.ID != 1 || item.Name != "learn Go" {
		t.Fatalf("created item = %+v", item)
	}

	list := httptest.NewRequest(http.MethodGet, "/items", nil)
	listed := httptest.NewRecorder()
	handler.ServeHTTP(listed, list)
	if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "learn Go") {
		t.Fatalf("list response = %d %s", listed.Code, listed.Body.String())
	}

	get := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	got := httptest.NewRecorder()
	handler.ServeHTTP(got, get)
	if got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "learn Go") {
		t.Fatalf("get response = %d %s", got.Code, got.Body.String())
	}

	update := httptest.NewRequest(http.MethodPut, "/items/1", strings.NewReader(`{"name":"learn HTTP"}`))
	update.Header.Set("Content-Type", "application/json")
	updated := httptest.NewRecorder()
	handler.ServeHTTP(updated, update)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), "learn HTTP") {
		t.Fatalf("update response = %d %s", updated.Code, updated.Body.String())
	}

	remove := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	removed := httptest.NewRecorder()
	handler.ServeHTTP(removed, remove)
	if removed.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", removed.Code, http.StatusNoContent)
	}
}

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	newHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK || recorder.Body.String() != `{"status":"ok"}
` {
		t.Fatalf("health response = %d %q", recorder.Code, recorder.Body.String())
	}
}
