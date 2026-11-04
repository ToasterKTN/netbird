//go:build !openapivalidation

package v1alpha1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAPIV1DisabledWithoutValidation(t *testing.T) {
	root := mux.NewRouter()
	sub := root.PathPrefix("/api/v1alpha1").Subrouter()
	if _, err := NewAPIV1Handler(context.Background(), sub, nil, nil, nil, nil); err != nil {
		t.Fatalf("handler creation must not fail: %v", err)
	}
	rec := httptest.NewRecorder()
	root.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1alpha1/peers", nil))
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", rec.Code)
	}
}
