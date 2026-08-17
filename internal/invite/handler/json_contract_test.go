package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestBodyRejectsTrailingDocument(t *testing.T) {
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"code":"AAAA"}{"code":"BBBB"}`))
	var target map[string]any
	if err := decode(req, &target); err == nil {
		t.Fatal("expected a request containing two JSON documents to be rejected")
	}
}
