package http_test

import (
	"context"
	"io"
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	s.AssertExpectations()

	_ = res.Body.Close()
}

func TestServer_HandlesExpectationWithQuery(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path?p=some%2Fpath")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path?p=some%2Fpath", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	s.AssertExpectations()

	_ = res.Body.Close()
}

func TestServer_HandlesAnythingMethodExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(httptest.Anything, "/test/path")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	s.AssertExpectations()

	_ = res.Body.Close()
}

func TestServer_HandlesAnythingPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, httptest.Anything)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	s.AssertExpectations()

	_ = res.Body.Close()
}

func TestServer_HandlesWildcardPathExpectation(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/*")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	s.AssertExpectations()

	_ = res.Body.Close()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	_ = resp.Body.Close()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	resp, _ := http.DefaultClient.Do(req)

	_ = resp.Body.Close()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path?p=somethingelse", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	s.AssertExpectations()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	s.AssertExpectations()
}

func TestServer_ExpectationReturnsBodyBytes(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Returns(400, []byte("test"))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := io.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsBodyString(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").ReturnsString(400, "test")

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := io.ReadAll(res.Body)
	assert.Equal(t, []byte("test"), b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsStatusCode(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").ReturnsStatus(400)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)
	b, _ := io.ReadAll(res.Body)
	assert.Empty(t, b)

	_ = res.Body.Close()
}

func TestServer_ExpectationReturnsHeaders(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Header("foo", "bar").ReturnsStatus(200)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	v := res.Header.Get("Foo")
	assert.Equal(t, "bar", v)

	_ = res.Body.Close()
}

func TestServer_ExpectationUsesHandleFunc(t *testing.T) {
	s := httptest.NewServer(t)
	t.Cleanup(s.Close)

	s.On(http.MethodGet, "/test/path").Handle(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/test/path", nil)
	require.NoError(t, err)
	res, err := http.DefaultClient.Do(req)

	require.NoError(t, err)
	assert.Equal(t, 400, res.StatusCode)

	_ = res.Body.Close()
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

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.URL()+"/", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)

	require.NoError(t, err)

	s.AssertExpectations()

	_ = resp.Body.Close()
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
