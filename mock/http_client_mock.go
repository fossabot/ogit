package mock

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// HTTPClient provides a simple way of setting up mocks for requests to an HTTP.
// The mocking is done by providing a RoundTripper implementation that
// intercepts the requests to HTTP service, passes them through a gorilla.Mux, and
// returns the response. The gorilla.Mux allows callers to conveniently set up
// mocked responses for a given HTTP endpoint.
type HTTPClient struct {
	roundTripper *roundTripperMocked
	// After the HTTP routes have been finalized, subsequent calls to register
	// new routes (via Mock()) will panic
	routesFinalized bool
}

// roundTripperMocked is a mocked RoundTripper for mocking responses from external services
type roundTripperMocked struct {
	mux *mux.Router
}

// NewHTTPClient returns an HTTPClient with no routes defined
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		roundTripper: &roundTripperMocked{mux.NewRouter()},
	}
}

// Client returns the mocked HTTP Client instance
func (hc *HTTPClient) Client() *http.Client {
	if !hc.routesFinalized {
		// Register a wildcard route to warn for calls to unknown routes
		hc.roundTripper.mux.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("\033[38;5;200mWARNING: No route registered for %s %s\033[0m", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotImplemented)
		})
		hc.routesFinalized = true
	}
	return &http.Client{Transport: hc.roundTripper}
}

// RoundTrip implements the RoundTrip method of RoundTripper for mocking purposes
func (rt roundTripperMocked) RoundTrip(r *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	rt.mux.ServeHTTP(w, r)
	return w.Result(), nil
}

// Mock registers an HTTP handler for the provided method and path
func (hc *HTTPClient) Mock(method, path string, handler http.HandlerFunc) *HTTPClient {
	if hc.routesFinalized {
		panic("http_client: Mock() cannot be called after Client()")
	}

	hc.roundTripper.mux.HandleFunc(path, handler).Methods(method)
	return hc
}
