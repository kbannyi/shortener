package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kbannyi/shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (s *MockService) Create(value string) (ID string) {
	return "mockid"
}

func (s *MockService) Get(ID string) (string, bool) {
	return "redirect", ID == "mockid"
}

func TestURLRouter(t *testing.T) {
	tests := []struct {
		method       string
		request      string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method:       http.MethodPost,
			request:      "/",
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodGet,
			request:      "/unknownid",
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodGet,
			request:      "/mockid",
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			method:       http.MethodPost,
			request:      "/",
			body:         "ya.ru",
			expectedBody: "http://localhost:8080/mockid",
			expectedCode: http.StatusCreated,
		},
		{
			method:       http.MethodPut,
			request:      "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			method:       http.MethodDelete,
			request:      "/",
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.request, strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			router := NewURLRouter(&MockService{}, config.Flags{
				RedirectBaseAddr: "http://localhost:8080/",
			})

			router.ServeHTTP(w, r)
			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}
