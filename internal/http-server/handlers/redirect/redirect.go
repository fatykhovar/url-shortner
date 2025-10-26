package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	resp "github.com/fatykhovar/url-shortner/internal/lib/api/response"
	"github.com/fatykhovar/url-shortner/internal/lib/logger/sl"
	"github.com/fatykhovar/url-shortner/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", requestid.Get(c)),
		)

		alias := c.Param("alias")
		if alias == "" {
			log.Info("alias is empty")
			c.JSON(http.StatusBadRequest, resp.Error("invalid request"))
			return
		}

		resURL, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrURLNotFound) {
			log.Info("url not found", slog.String("alias", alias))

			c.JSON(http.StatusNotFound, resp.Error("not found"))

			return
		}
		if err != nil {
			log.Error("failed to get url", sl.Err(err))

			c.JSON(http.StatusInternalServerError, resp.Error("internal server error"))

			return
		}

		log.Info("url found", slog.String("alias", alias), slog.String("url", resURL))

		c.Redirect(http.StatusFound, resURL)
	}
}
