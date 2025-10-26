package redirect_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/redirect"
	"github.com/fatykhovar/url-shortner/internal/http-server/handlers/redirect/mocks"
	"github.com/fatykhovar/url-shortner/internal/lib/api"
	"github.com/fatykhovar/url-shortner/internal/lib/logger/handlers/slogdiscard"
)

func TestGetHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "example",
			url:   "https://www.example.com/",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			router := gin.Default()
			handler := redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock)
			router.GET("/:alias", handler)

			ts := httptest.NewServer(router)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			// Check the final URL after redirection.
			assert.Equal(t, tc.url, redirectedToURL)
		})
	}
}
