package save

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	resp "github.com/fatykhovar/url-shortner/internal/lib/api/response"
	"github.com/fatykhovar/url-shortner/internal/lib/logger/sl"
	"github.com/fatykhovar/url-shortner/internal/lib/random"
	"github.com/fatykhovar/url-shortner/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

// TODO: move to config if needed
const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) gin.HandlerFunc {
	return func(c *gin.Context) {
		const op = "handlers.url.save.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", requestid.Get(c)),
		)

		var req Request

		err := c.ShouldBindJSON(&req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			c.JSON(http.StatusBadRequest, resp.Error("empty request"))

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			c.JSON(http.StatusBadRequest, resp.Error("failed to decode request body"))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			c.JSON(http.StatusBadRequest, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		if req.URL == "" {
			log.Error("url is empty")

			c.JSON(http.StatusBadRequest, resp.Error("field URL is a required field"))

			return
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			c.JSON(http.StatusConflict, resp.Error("url already exists"))

			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))

			c.JSON(http.StatusInternalServerError, resp.Error("failed to add url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		c.JSON(http.StatusOK, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}
