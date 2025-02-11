package del

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"golang.org/x/exp/slog"

	resp "github.com/nedocodil/microservices/url-shortener/internal/lib/api/response"
	"github.com/nedocodil/microservices/url-shortener/internal/lib/logger/sl"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.del.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")

			render.JSON(w, r, resp.Error("invalid request"))

			return
		}

		err := urlDeleter.DeleteURL(alias)
		if err != nil {
			log.Error("failed to delete url", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to delete url"))

			return
		}

		log.Info("url deleted")



		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias: alias,
		})
	}
}