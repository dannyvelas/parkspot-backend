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
	*http.Server
}

func NewServer(c config.Config, app app.App) Server {
	router := newRouter(c, app)

	server := &http.Server{
		Addr:         ":" + c.Http.Port,
		Handler:      router,
		ReadTimeout:  c.Http.ReadTimeout,
		WriteTimeout: c.Http.WriteTimeout,
		IdleTimeout:  c.Http.IdleTimeout,
	}

	return Server{
		server,
	}
}

func newRouter(c config.Config, app app.App) (router *chi.Mux) {
	router = chi.NewRouter()

	// middlewares
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   c.Http.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		MaxAge:           300,
	}))

	// handlers
	middleware := newMiddleware(app.JWTService)
	authHandler := newAuthHandler(c.Http, app.JWTService, app.AuthService)
	residentHandler := newResidentHandler(app.ResidentService)
	visitorHandler := newVisitorHandler(app.VisitorService)
	carHandler := newCarHandler(app.CarService)
	permitHandler := newPermitHandler(app.PermitService)

	// index
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, "hello world")
	}))

	// api
	router.Route("/api", func(r chi.Router) {
		r.Group(func(anyoneRouter chi.Router) {
			anyoneRouter.Post("/login", authHandler.login())
			anyoneRouter.Post("/logout", authHandler.logout())
			anyoneRouter.Post("/refresh-tokens", authHandler.refreshTokens()) // needs to be here instead of userRouter. this is because user-router checks access tokens and an access token might be expired when this is called
			anyoneRouter.Post("/password-reset-email", authHandler.sendResetPasswordEmail())
		})

		r.Group(func(adminRouter chi.Router) {
			adminRouter.Use(middleware.authenticate(models.AdminRole))
			adminRouter.Delete("/permit/{id:[0-9]+}", permitHandler.deleteOne())
			adminRouter.Delete("/resident/{id}", residentHandler.deleteOne())
			adminRouter.Post("/resident", residentHandler.create())
			adminRouter.Put("/resident", residentHandler.edit())
			adminRouter.Put("/permit", permitHandler.edit())
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(middleware.authenticate(models.AdminRole, models.SecurityRole))
			officeRouter.Get("/residents", residentHandler.getAll())
			officeRouter.Get("/resident/{id}", residentHandler.getOne())
			officeRouter.Get("/car/{id}", carHandler.getOne())
		})

		r.Group(func(adminAndResidentRouter chi.Router) {
			adminAndResidentRouter.Use(middleware.authenticate(models.AdminRole, models.ResidentRole))
			adminAndResidentRouter.Post("/permit", permitHandler.create())
			adminAndResidentRouter.Post("/car", carHandler.create())
			adminAndResidentRouter.Put("/car", carHandler.edit())
			adminAndResidentRouter.Delete("/car/{id}", carHandler.deleteOne())
		})

		r.Group(func(residentRouter chi.Router) {
			residentRouter.Use(middleware.authenticate(models.ResidentRole))
			residentRouter.Post("/visitor", visitorHandler.create())
			residentRouter.Delete("/visitor/{id}", visitorHandler.deleteOne())
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(middleware.authenticate(models.AdminRole, models.SecurityRole, models.ResidentRole))
			userRouter.Put("/user/password", authHandler.resetPassword())
			userRouter.Get("/hello", sayHello())
			userRouter.Get("/permits/all", permitHandler.get(models.AnyStatus))
			userRouter.Get("/permits/active", permitHandler.get(models.ActiveStatus))
			userRouter.Get("/permits/exceptions", permitHandler.get(models.ExceptionStatus))
			userRouter.Get("/permits/expired", permitHandler.get(models.ExpiredStatus))
			userRouter.Get("/permit/{id:[0-9]+}", permitHandler.getOne())
			userRouter.Get("/visitors/active", visitorHandler.get(models.ActiveStatus))
			userRouter.Get("/resident/{id}/cars", carHandler.getOfResident())
			userRouter.Get("/cars", carHandler.get())
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
