package http_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	httptest "github.com/hamba/testutils/http"
	"github.com/stretchr/testify/assert"
)

func TestServer_HandlesExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path")

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesAnythingMethodExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On(httptest.Anything, "/test/path")

	res, err := http.Post(s.URL()+"/test/path", "text/plain", bytes.NewReader([]byte{}))
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesAnythingPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", httptest.Anything)

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesWildcardPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/*")

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesUnexpectedMethodRequest(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if !mockT.Failed() {
			t.Error("Expected error when no expectation on request")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()

	s.On("POST", "/")

	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesUnexpectedPathRequest(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if !mockT.Failed() {
			t.Error("Expected error when no expectation on request")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()
	s.On("GET", "/foobar")

	s.On("GET", "/")

	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesExpectationNTimes(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if !mockT.Failed() {
			t.Error("Expected error when expectation times used")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()
	s.On("GET", "/test/path").Times(2)

	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesExpectationUnlimitedTimes(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if mockT.Failed() {
			t.Error("Unexpected error on request")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()
	s.On("GET", "/test/path")

	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_ExpectationReturnsBodyBytes(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path").Returns(400, []byte("test"))

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsBodyString(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path").ReturnsString(400, "test")

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsStatusCode(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path").ReturnsStatus(400)

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Len(t, b, 0)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsHeaders(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path").Header("foo", "bar").ReturnsStatus(200)

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	v := res.Header.Get("foo")
	assert.Equal(t, "bar", v)

	_ = res.Body.Close()
}

func TestServer_ExpectationUsesHandleFunc(t *testing.T) {
	s := httptest.NewServer(t)
	defer s.Close()

	s.On("GET", "/test/path").Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})

	res, err := http.Get(s.URL() + "/test/path")
	assert.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
}

func TestServer_AssertExpectationsOnUnlimited(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if !mockT.Failed() {
			t.Error("Expected error when asserting expectations")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()
	s.On("POST", "/")

	s.AssertExpectations()
}

func TestServer_AssertExpectationsOnNTimes(t *testing.T) {
	mockT := new(testing.T)
	defer func() {
		if !mockT.Failed() {
			t.Error("Expected error when asserting expectations")
		}

	}()

	s := httptest.NewServer(mockT)
	defer s.Close()
	s.On("POST", "/").Times(1)

	s.AssertExpectations()
}
