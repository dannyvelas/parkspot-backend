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

	router.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Post("/login", Login(jwtMiddleware, adminRepo))

		apiRouter.Route("/", func(adminRouter chi.Router) {
			adminRouter.Use(jwtMiddleware.Authenticate) // jwtMiddleware.AuthenticateOffice (admin/security)
			adminRouter.Get("/permits/active", getActive(permitRepo))
			adminRouter.Get("/permits", getAll(permitRepo))
			adminRouter.Get("/permits/exceptions", getExceptions(permitRepo))
			adminRouter.Get("/permits/expired", getExpired(permitRepo))
			adminRouter.Get("/residents", getAllResidents(residentRepo))
		})

		//apiRouter.Use(jwtMiddleware.AuthenticateUser) (admin/security/resident)
		apiRouter.Get("/hello", sayHello())
		apiRouter.Post("/permit", create(permitRepo, carRepo, residentRepo, dateFormat))
		//apiRouter.Get("/permit/{id:[0-9]+}", getOnePermit(permitRepo))
		apiRouter.Get("/car/{id}", getOneCar(carRepo))
		apiRouter.Put("/car/{id}", editCar(carRepo))
		apiRouter.Delete("/permit/{id:[0-9]+}", deletePermit(permitRepo))
	})

	return
}
