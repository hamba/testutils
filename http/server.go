/*
Package http provides utilities for HTTP testing.

Example Server Usage:

	import (
		"net/http"
		"testing"

		httptest "github.com/hamba/testutils/http"
	)

	func TestSomething(t *testing.T) {
		s := httptest.NewServer(t)
		s.On(http.MethodGet, "/your/endpoint").Times(2).ReturnsString(http.StatusOK, "some return")
		defer s.Close()

		// Call the server

		s.AssertExpectations()
	}
*/
package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/ryanuber/go-glob"
)

const (
	// Anything is used where the expectation should not be considered.
	Anything = "httptest.Anything"
)

// Expectation represents an http request expectation.
type Expectation struct {
	method string
	path   string
	qry    *url.Values

	fn http.HandlerFunc

	headers []string
	body    []byte
	status  int

	times  int
	called int
}

// Times sets the number of times the request can be made.
func (e *Expectation) Times(times int) *Expectation {
	e.times = times
	e.called = times

	return e
}

// Header sets the HTTP headers that should be returned.
func (e *Expectation) Header(k, v string) *Expectation {
	e.headers = append(e.headers, k, v)

	return e
}

// Handle sets the HTTP handler function to be run on the request.
func (e *Expectation) Handle(fn http.HandlerFunc) {
	e.fn = fn
}

// ReturnsStatus sets the HTTP stats code to return.
func (e *Expectation) ReturnsStatus(status int) {
	e.body = []byte{}
	e.status = status
}

// Returns sets the HTTP stats and body bytes to return.
func (e *Expectation) Returns(status int, body []byte) {
	e.body = body
	e.status = status
}

// ReturnsString sets the HTTP stats and body string to return.
func (e *Expectation) ReturnsString(status int, body string) {
	e.body = []byte(body)
	e.status = status
}

// Server represents a mock http server.
type Server struct {
	t   *testing.T
	srv *httptest.Server

	expect []*Expectation
}

// NewServer creates a new mock http server.
func NewServer(t *testing.T) *Server {
	t.Helper()

	srv := &Server{
		t: t,
	}
	srv.srv = httptest.NewServer(http.HandlerFunc(srv.handler))

	return srv
}

// URL returns the url of the mock server.
func (s *Server) URL() string {
	return s.srv.URL
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	qry := r.URL.Query()
	for i, exp := range s.expect {
		if exp.method != method && exp.method != Anything {
			continue
		}

		if exp.path != Anything && !glob.Glob(exp.path, path) {
			continue
		}

		if exp.qry != nil {
			found := false
			for k, v := range *exp.qry {
				if !qry.Has(k) {
					break
				}
				if elementsMatch(v, qry[k]) {
					found = true
				}
			}
			if !found {
				continue
			}
		}

		for j := 0; j < len(exp.headers); j += 2 {
			w.Header().Add(exp.headers[j], exp.headers[j+1])
		}

		if exp.fn != nil {
			exp.fn(w, r)
		} else {
			w.WriteHeader(exp.status)
			if len(exp.body) > 0 {
				_, _ = w.Write(exp.body)
			}
		}

		exp.called--
		if exp.called == 0 {
			s.expect = append(s.expect[:i], s.expect[i+1:]...)
		}
		return
	}

	s.t.Errorf("Unexpected call to %s %s", method, path)
}

// On creates an expectation of a request on the server.
func (s *Server) On(method, path string) *Expectation {
	var qry *url.Values
	if parts := strings.SplitN(path, "?", 2); len(parts) == 2 {
		path = parts[0]
		if val, err := url.ParseQuery(parts[1]); err == nil {
			qry = &val
		}
	}

	exp := &Expectation{
		method: method,
		path:   path,
		qry:    qry,
		times:  -1,
		called: -1,
		status: 200,
	}
	s.expect = append(s.expect, exp)

	return exp
}

// AssertExpectations asserts all expectations have been met.
func (s *Server) AssertExpectations() {
	for _, exp := range s.expect {
		var call string
		if exp.method != Anything {
			call = exp.method
		}
		if exp.path != Anything {
			if call != "" {
				call += " "
			}
			call += exp.path
		}
		if exp.qry != nil {
			if call != "" || exp.path == Anything {
				call += " "
			}
			call += exp.qry.Encode()
		}

		switch exp.called {
		case -1:
			s.t.Errorf("mock: server: Expected a call to %s but got none", call)
		case 0:
		default:
			s.t.Errorf("mock: server: Expected a call to %s %d times but got called %d times", call, exp.times, exp.times-exp.called)
		}
	}
}

// Close closes the server.
func (s *Server) Close() {
	s.srv.Close()
}

func elementsMatch(a, b []string) bool {
	aLen := len(a)
	bLen := len(b)

	visited := make([]bool, bLen)
	for i := 0; i < aLen; i++ {
		found := false
		element := a[i]
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if element == b[j] {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
