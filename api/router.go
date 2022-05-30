package api

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func NewRouter(httpConfig config.HttpConfig,
	tokenConfig config.TokenConfig,
	dateFormat string,
	adminRepo storage.AdminRepo,
	permitRepo storage.PermitRepo,
	carRepo storage.CarRepo,
	residentRepo storage.ResidentRepo,
) (router *chi.Mux) {
	router = chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   httpConfig.CORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	jwtMiddleware := NewJWTMiddleware(tokenConfig)

	router.Route("/api", func(r chi.Router) {
		r.Group(func(anyoneRouter chi.Router) {
			anyoneRouter.Post("/login", Login(jwtMiddleware, adminRepo))
			anyoneRouter.Post("/logout", Logout())
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(jwtMiddleware.Authenticate(AdminRole)) //, SecurityRole
			officeRouter.Get("/permits/active", getActive(permitRepo))
			officeRouter.Get("/permits", getAll(permitRepo))
			officeRouter.Get("/permits/exceptions", getExceptions(permitRepo))
			officeRouter.Get("/permits/expired", getExpired(permitRepo))
			officeRouter.Get("/residents", getAllResidents(residentRepo))
			officeRouter.Get("/resident/{id}", getOneResident(residentRepo))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(jwtMiddleware.Authenticate(AdminRole, ResidentRole)) //, SecurityRole
			userRouter.Get("/hello", sayHello())
			userRouter.Post("/permit", create(permitRepo, carRepo, residentRepo, dateFormat))
			//userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(permitRepo))
			userRouter.Get("/car/{id}", getOneCar(carRepo))
			userRouter.Put("/car/{id}", editCar(carRepo))
			userRouter.Delete("/permit/{id:[0-9]+}", deletePermit(permitRepo))
			userRouter.Get("/permits/search", searchPermits(permitRepo))
		})
	})

	return
}
