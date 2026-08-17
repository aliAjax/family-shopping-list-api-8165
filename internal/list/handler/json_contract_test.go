package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyRejectsTrailingDocument(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"weekly"}{"name":"later"}`))
	var target map[string]any
	if err := decode(req, &target); err == nil {
		t.Fatal("expected a request containing two JSON documents to be rejected")
	}
}
