package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateKeyHandlerMissingAuthorization(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	app := &App{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/validate-key", nil)
	res := httptest.NewRecorder()

	app.validateKeyHandler(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}

	if !strings.Contains(res.Body.String(), "Authorization header não encontrado") {
		t.Fatalf("unexpected response body: %s", res.Body.String())
	}
}

func TestValidateKeyHandlerInvalidKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	app := &App{DB: db}

	key := "invalid-key"
	keyHash := hashAPIKey(key)

	mock.ExpectQuery(`SELECT id FROM api_keys WHERE key_hash = \$1 AND is_active = true`).
		WithArgs(keyHash).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/validate-key", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	res := httptest.NewRecorder()

	app.validateKeyHandler(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}

	if !strings.Contains(res.Body.String(), "Chave de API inválida ou inativa") {
		t.Fatalf("unexpected response body: %s", res.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}

func TestValidateKeyHandlerValidKey(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	app := &App{DB: db}

	key := "valid-key"
	keyHash := hashAPIKey(key)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`SELECT id FROM api_keys WHERE key_hash = \$1 AND is_active = true`).
		WithArgs(keyHash).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/validate-key", nil)
	req.Header.Set("Authorization", "Bearer "+key)

	res := httptest.NewRecorder()

	app.validateKeyHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}

	if !strings.Contains(res.Body.String(), "Chave válida") {
		t.Fatalf("unexpected response body: %s", res.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet database expectations: %v", err)
	}
}
