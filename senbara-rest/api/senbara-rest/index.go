package senbaraRest

//go:generate tar czf code.tar.gz --exclude .git -C ../../.. .

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/pojntfx/senbara/senbara-common/pkg/persisters"
	"github.com/pojntfx/senbara/senbara-rest/pkg/api"
	"github.com/pojntfx/senbara/senbara-rest/pkg/controllers"
)

//go:embed code.tar.gz
var code []byte

var (
	p *persisters.Persister
	c *controllers.Controller
	s *openapi3.T
)

func SenbaraRESTHandler(
	w http.ResponseWriter,
	r *http.Request,
	c *controllers.Controller,
	s *openapi3.T,
) {
	mux := http.NewServeMux()

	mux.Handle(
		"/",
		middleware.OapiRequestValidatorWithOptions(
			s,
			&middleware.Options{
				Options: openapi3filter.Options{
					AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
						_, err := c.Authenticate(r)

						return err
					},
				},
			},
		)(
			api.Handler(
				api.NewStrictHandler(c, []api.StrictMiddlewareFunc{
					c.Authorize,
				}),
			),
		),
	)

	mux.ServeHTTP(w, r)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = r.URL.Query().Get("path")

	opts := &slog.HandlerOptions{}
	if os.Getenv("VERBOSE") == "true" {
		opts.Level = slog.LevelDebug
	}
	log := slog.New(slog.NewJSONHandler(os.Stderr, opts))

	if p == nil {
		p = persisters.NewPersister(slog.New(log.Handler().WithGroup("persister")), os.Getenv("POSTGRES_URL"))

		if err := p.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	if s == nil {
		var err error
		s, err = api.GetSwagger()
		if err != nil {
			panic(err)
		}
	}

	if c == nil {
		c = controllers.NewController(
			slog.New(log.Handler().WithGroup("controller")),

			p,

			s,

			os.Getenv("OIDC_ISSUER"),
			os.Getenv("OIDC_CLIENT_ID"),
			os.Getenv("OIDC_REDIRECT_URL"),

			os.Getenv("PRIVACY_URL"),
			os.Getenv("IMPRINT_URL"),
		)

		if err := c.Init(r.Context()); err != nil {
			panic(err)
		}
	}

	SenbaraRESTHandler(w, r, c, s)
}
