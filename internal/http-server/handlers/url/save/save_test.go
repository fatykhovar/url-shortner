package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	// "github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/url/save"
	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/url/save/mocks"
	"github.com/fatykhovar/url-shortner/internal/lib/logger/handlers/slogdiscard"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		url          string
		respError    string
		mockError    error
		expectedCode int
	}{
		{
			name:         "Success",
			alias:        "test_alias",
			url:          "https://google.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty alias",
			alias:        "",
			url:          "https://google.com",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Empty URL",
			url:          "",
			alias:        "some_alias",
			respError:    "field URL is a required field",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid URL",
			url:          "some invalid URL",
			alias:        "some_alias",
			respError:    "field URL is not a valid URL",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "SaveURL Error",
			alias:        "test_alias",
			url:          "https://google.com",
			respError:    "failed to add url",
			mockError:    errors.New("unexpected error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockError).
					Once()
			}

			router := gin.Default()
			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)
			router.POST("/url", handler)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tc.expectedCode)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
