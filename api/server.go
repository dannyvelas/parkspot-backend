package api

import (
	"context"
	"github.com/dannyvelas/lasvistas_api/app"
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type Server struct {
	http.Server
}

func NewServer(c config.Config, app app.App) Server {
	router := newRouter(c, app)

	server := http.Server{
		Addr:         ":" + c.Http().Port(),
		Handler:      router,
		ReadTimeout:  c.Http().ReadTimeout(),
		WriteTimeout: c.Http().WriteTimeout(),
		IdleTimeout:  c.Http().IdleTimeout(),
	}

	return Server{
		server,
	}
}

func newRouter(c config.Config, app app.App) (router *chi.Mux) {
	router = chi.NewRouter()

	// middlewares
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   c.Http().CORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		MaxAge:           300,
	}))

	// handlers
	middleware := NewMiddleware(app.JWTService)
	authHandler := NewAuthHandler(app.JWTService, app.AuthService)
	residentHandler := NewResidentHandler(app.ResidentService)

	// index
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, "hello world")
	}))

	// api
	router.Route("/api", func(r chi.Router) {
		r.Group(func(anyoneRouter chi.Router) {
			anyoneRouter.Post("/login-resident", authHandler.login())
			anyoneRouter.Post("/login-admin", authHandler.login())
			anyoneRouter.Post("/logout", authHandler.logout())
			anyoneRouter.Post("/refresh-tokens", authHandler.refreshTokens())
			anyoneRouter.Post("/password-reset-email", authHandler.sendResetPasswordEmail())
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(middleware.authenticate(models.AdminRole)) //, SecurityRole
			//officeRouter.Delete("/permit/{id:[0-9]+}", deletePermit(repos.Permit, repos.Resident, repos.Car))
			officeRouter.Get("/residents", residentHandler.GetAll())
			officeRouter.Get("/resident/{id}", residentHandler.GetOne())
			officeRouter.Post("/account", residentHandler.Create())
			officeRouter.Delete("/resident/{id}", residentHandler.Delete())
			officeRouter.Put("/resident/{id}", residentHandler.Edit())
			//officeRouter.Get("/car/{id}", getOneCar(repos.Car))
			//officeRouter.Put("/car/{id}", editCar(repos.Car))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(middleware.authenticate(models.AdminRole, models.ResidentRole)) //, SecurityRole
			userRouter.Get("/hello", sayHello())
			//userRouter.Get("/permits/all", getPermits(repos.Permit, models.AllPermits))
			//userRouter.Get("/permits/active", getPermits(repos.Permit, models.ActivePermits))
			//userRouter.Get("/permits/exceptions", getPermits(repos.Permit, models.ExceptionPermits))
			//userRouter.Get("/permits/expired", getPermits(repos.Permit, models.ExpiredPermits))
			//userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(repos.Permit))
			//userRouter.Post("/permit", createPermit(repos.Permit, repos.Resident, repos.Car, config.DateFormat))
			//userRouter.Get("/visitors", getActiveVisitors(repos.Visitor))
			//userRouter.Put("/account/password", resetPassword(app.JWTService, app.AuthService))
		})

		r.Group(func(residentRouter chi.Router) {
			residentRouter.Use(middleware.authenticate(models.ResidentRole))
			//residentRouter.Post("/visitor", createVisitor(repos.Visitor))
			//residentRouter.Delete("/visitor/{id}", deleteVisitor(repos.Visitor))
		})
	})

	return
}

func (s Server) Start(errChannel chan<- error) {
	log.Info().Msgf("Server started on: %s", s.Addr)
	errChannel <- s.ListenAndServe()
}

func (s Server) ShutdownGracefully(timeout time.Duration) {
	log.Info().Msg("Gracefully shutting down...")

	gracefullCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.Shutdown(gracefullCtx); err != nil {
		log.Error().Msgf("Error shutting down the server: %v", err)
	} else {
		log.Info().Msg("HttpServer gracefully shut down")
	}
}
