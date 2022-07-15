package http_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	httptest "github.com/hamba/testutils/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_HandlesExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path")

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesExpectationWithQuery(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path?p=some%2Fpath")

	res, err := http.Get(s.URL() + "/test/path?p=some%2Fpath")
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesAnythingMethodExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(httptest.Anything, "/test/path")

	res, err := http.Post(s.URL()+"/test/path", "text/plain", bytes.NewReader([]byte{}))
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesAnythingPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, httptest.Anything)

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesWildcardPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/*")

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}

func TestServer_HandlesUnexpectedMethodRequest(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when no expectation on request")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)

	s.On("POST", "/")

	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesUnexpectedPathRequest(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when no expectation on request")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodGet, "/foobar")

	s.On(http.MethodGet, "/")

	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesUnexpectedPathQueryRequest(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when no expectation on request")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodGet, "/test/path?a=other")
	s.On(http.MethodGet, "/test/path?p=something")

	_, _ = http.Get(s.URL() + "/test/path?p=somethingelse")
}

func TestServer_HandlesExpectationNTimes(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when expectation times used")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodGet, "/test/path").Times(2)

	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_HandlesExpectationUnlimitedTimes(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if mockT.Failed() {
			t.Error("Unexpected error on request")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodGet, "/test/path")

	_, _ = http.Get(s.URL() + "/test/path")
	_, _ = http.Get(s.URL() + "/test/path")
}

func TestServer_ExpectationReturnsBodyBytes(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Returns(400, []byte("test"))

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsBodyString(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").ReturnsString(400, "test")

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsStatusCode(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").ReturnsStatus(400)

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Len(t, b, 0)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsHeaders(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Header("foo", "bar").ReturnsStatus(200)

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	v := res.Header.Get("foo")
	assert.Equal(t, "bar", v)

	_ = res.Body.Close()
}

func TestServer_ExpectationUsesHandleFunc(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
	})

	res, err := http.Get(s.URL() + "/test/path")
	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
}

func TestServer_AssertExpectations(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if mockT.Failed() {
			t.Error("Expected no error when asserting expectations")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodGet, "/").Times(1)

	_, err := http.Get(s.URL() + "/")
	assert.NoError(t, err)

	s.AssertExpectations()
}

func TestServer_AssertExpectationsOnUnlimited(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when asserting expectations")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodPost, "/")

	s.AssertExpectations()
}

func TestServer_AssertExpectationsOnNTimes(t *testing.T) {
	mockT := new(testing.T)
	t.Cleanup(func() {
		if !mockT.Failed() {
			t.Error("Expected error when asserting expectations")
		}
	})

	s := httptest.NewServer(mockT)
	t.Cleanup(s.Close)
	s.On(http.MethodPost, "/").Times(1)

	s.AssertExpectations()
}
