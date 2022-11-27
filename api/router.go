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
			anyoneRouter.Post("/login-resident", login[models.Resident](jwtMiddleware, repos.Resident))
			anyoneRouter.Post("/login-admin", login[models.Admin](jwtMiddleware, repos.Admin))
			anyoneRouter.Post("/logout", logout())
			anyoneRouter.Post("/refresh-tokens", refreshTokens(jwtMiddleware, repos.Admin, repos.Resident))
			anyoneRouter.Post("/password-reset-email", sendResetPasswordEmail(jwtMiddleware, httpConfig, oauthConfig, repos.Admin, repos.Resident))
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(jwtMiddleware.authenticate(models.AdminRole)) //, SecurityRole
			officeRouter.Delete("/permit/{id:[0-9]+}", deletePermit(repos.Permit, repos.Resident, repos.Car))
			officeRouter.Get("/residents", getAllResidents(repos.Resident))
			officeRouter.Get("/resident/{id}", getOneResident(repos.Resident))
			officeRouter.Post("/account", createResident(repos.Resident))
			officeRouter.Delete("/resident/{id}", deleteResident(repos.Resident))
			officeRouter.Put("/resident/{id}", editResident(repos.Resident))
			officeRouter.Get("/car/{id}", getOneCar(repos.Car))
			officeRouter.Put("/car/{id}", editCar(repos.Car))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(jwtMiddleware.authenticate(models.AdminRole, models.ResidentRole)) //, SecurityRole
			userRouter.Get("/hello", sayHello())
			userRouter.Get("/permits/all", getPermits(repos.Permit, models.AllPermits))
			userRouter.Get("/permits/active", getPermits(repos.Permit, models.ActivePermits))
			userRouter.Get("/permits/exceptions", getPermits(repos.Permit, models.ExceptionPermits))
			userRouter.Get("/permits/expired", getPermits(repos.Permit, models.ExpiredPermits))
			userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(repos.Permit))
			userRouter.Post("/permit", createPermit(repos.Permit, repos.Resident, repos.Car, dateFormat))
			userRouter.Get("/visitors", getActiveVisitors(repos.Visitor))
			userRouter.Put("/account/password", resetPassword(jwtMiddleware, repos.Admin, repos.Resident))
		})

		r.Group(func(residentRouter chi.Router) {
			residentRouter.Use(jwtMiddleware.authenticate(models.ResidentRole))
			residentRouter.Post("/visitor", createVisitor(repos.Visitor))
			residentRouter.Delete("/visitor/{id}", deleteVisitor(repos.Visitor))
		})
	})

	return
}
