package save

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/exp/slog"

	resp "github.com/nedocodil/microservices/url-shortener/internal/lib/api/response"
	"github.com/nedocodil/microservices/url-shortener/internal/lib/logger/sl"
	"github.com/nedocodil/microservices/url-shortener/internal/lib/random"
	"github.com/nedocodil/microservices/url-shortener/internal/storage"
)

type Request struct {
	URL  string `json:"url" validate:"required,isValidURL"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias  string `json:"alias,omitempty"`
}

// TODO: move to config
const aliasLength = 6

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func isValidURL(fl validator.FieldLevel) bool {
	u, err := url.Parse(fl.Field().String())
	return err == nil && u.Scheme != "" && u.Host != ""
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))

			return 
		}

		log.Info("request body decoded", slog.Any("request", req))

		v := validator.New()
		v.RegisterValidation("isValidURL", isValidURL)

		if err := v.Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if errors.Is(err, storage.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))

			render.JSON(w, r, resp.Error("url already exists"))

			return
		}

		if err != nil {
			log.Error("failed to save url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save url"))

			return
		}

		log.Info("url added", slog.Int64("id", id))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}