package api

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
)

func NewRouter(
	httpConfig config.HttpConfig,
	tokenConfig config.TokenConfig,
	oauthConfig config.OAuthConfig,
	dateFormat string,
	repos storage.Repos,
) (router *chi.Mux) {
	router = chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   httpConfig.CORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		MaxAge:           300,
	}))

	jwtMiddleware := NewJWTMiddleware(tokenConfig)

	// index
	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, "hello world")
	}))

	// api
	router.Route("/api", func(r chi.Router) {
		r.Group(func(anyoneRouter chi.Router) {
			anyoneRouter.Post("/login", login(jwtMiddleware, repos.Admin, repos.Resident))
			anyoneRouter.Post("/logout", logout())
			anyoneRouter.Post("/refresh_tokens", refreshTokens(jwtMiddleware, repos.Admin, repos.Resident))
			anyoneRouter.Post("/password-reset-email", sendResetPasswordEmail(jwtMiddleware, httpConfig, oauthConfig, repos.Admin, repos.Resident))
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(jwtMiddleware.authenticate(AdminRole)) //, SecurityRole
			officeRouter.Get("/permits/all", getPermits(repos.Permit, models.AllPermits))
			officeRouter.Get("/permits/active", getPermits(repos.Permit, models.ActivePermits))
			officeRouter.Get("/permits/exceptions", getPermits(repos.Permit, models.ExceptionPermits))
			officeRouter.Get("/permits/expired", getPermits(repos.Permit, models.ExpiredPermits))
			officeRouter.Delete("/permit/{id:[0-9]+}", deletePermit(repos.Permit, repos.Resident, repos.Car))
			officeRouter.Get("/residents", getAllResidents(repos.Resident))
			officeRouter.Get("/resident/{id}", getOneResident(repos.Resident))
			officeRouter.Get("/visitors", getActiveVisitors(repos.Visitor))
			officeRouter.Get("/visitors/search", searchVisitors(repos.Visitor))
			officeRouter.Post("/account", createResident(repos.Resident))
			officeRouter.Delete("/resident/{id}", deleteResident(repos.Resident))
			officeRouter.Get("/car/{id}", getOneCar(repos.Car))
			officeRouter.Put("/car/{id}", editCar(repos.Car))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(jwtMiddleware.authenticate(AdminRole, ResidentRole)) //, SecurityRole
			userRouter.Put("/account/password", resetPassword(jwtMiddleware, repos.Admin, repos.Resident))
			userRouter.Get("/hello", sayHello())
			userRouter.Post("/permit", createPermit(repos.Permit, repos.Resident, repos.Car, dateFormat))
			userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(repos.Permit))
			userRouter.Get("/resident/{id}/permits", getAllPermitsOfResident(repos.Permit))
			userRouter.Get("/resident/{id}/permits/active", getActivePermitsOfResident(repos.Permit))
		})

		r.Group(func(residentRouter chi.Router) {
			residentRouter.Use(jwtMiddleware.authenticate(ResidentRole))
			residentRouter.Get("/me/visitors", getVisitorsOfResident(repos.Visitor))
			residentRouter.Post("/visitor", createVisitor(repos.Visitor))
			residentRouter.Delete("/visitor/{id}", deleteVisitor(repos.Visitor))
		})
	})

	return
}
