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
	adminRepo storage.AdminRepo,
	permitRepo storage.PermitRepo,
	carRepo storage.CarRepo,
	residentRepo storage.ResidentRepo,
	visitorRepo storage.VisitorRepo,
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
			anyoneRouter.Post("/login", login(jwtMiddleware, adminRepo, residentRepo))
			anyoneRouter.Post("/logout", logout())
			anyoneRouter.Post("/password-reset-email", sendResetPasswordEmail(jwtMiddleware, oauthConfig, adminRepo, residentRepo))
			anyoneRouter.Put("/account/password", resetPassword(jwtMiddleware, adminRepo, residentRepo))
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(jwtMiddleware.authenticate(AdminRole)) //, SecurityRole
			officeRouter.Get("/permits", getPermits(permitRepo, models.AllPermits))
			officeRouter.Get("/permits/active", getPermits(permitRepo, models.ActivePermits))
			officeRouter.Get("/permits/exceptions", getPermits(permitRepo, models.ExceptionPermits))
			officeRouter.Get("/permits/expired", getPermits(permitRepo, models.ExpiredPermits))
			officeRouter.Get("/permits/search", searchPermits(permitRepo))
			officeRouter.Delete("/permit/{id:[0-9]+}", deletePermit(permitRepo, residentRepo, carRepo))
			officeRouter.Get("/residents", getAllResidents(residentRepo))
			officeRouter.Get("/resident/{id}", getOneResident(residentRepo))
			officeRouter.Get("/visitors", getActiveVisitors(visitorRepo))
			officeRouter.Get("/visitors/search", searchVisitors(visitorRepo))
			officeRouter.Post("/account", createResident(residentRepo))
			officeRouter.Delete("/account/{id}", deleteResident(residentRepo))
			officeRouter.Get("/car/{id}", getOneCar(carRepo))
			officeRouter.Put("/car/{id}", editCar(carRepo))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(jwtMiddleware.authenticate(AdminRole, ResidentRole)) //, SecurityRole
			userRouter.Get("/hello", sayHello())
			userRouter.Post("/permit", createPermit(permitRepo, residentRepo, carRepo, dateFormat))
			userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(permitRepo))
			userRouter.Get("/resident/{id}/permits", getAllPermitsOfResident(permitRepo))
			userRouter.Get("/resident/{id}/permits/active", getActivePermitsOfResident(permitRepo))
		})

		r.Group(func(residentRouter chi.Router) {
			residentRouter.Use(jwtMiddleware.authenticate(ResidentRole))
			residentRouter.Get("/me/visitors", getVisitorsOfResident(visitorRepo))
			residentRouter.Post("/visitor", createVisitor(visitorRepo))
			residentRouter.Delete("/visitor/{id}", deleteVisitor(visitorRepo))
		})
	})

	return
}
