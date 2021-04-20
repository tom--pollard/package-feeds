package testutils

import (
	"net/http"
	"net/http/httptest"
)

type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func HTTPServerMock(handlerFuncs map[string]HTTPHandlerFunc) *httptest.Server {
	handler := http.NewServeMux()
	for endpoint, f := range handlerFuncs {
		handler.HandleFunc(endpoint, f)
	}
	srv := httptest.NewServer(handler)

	return srv
}
