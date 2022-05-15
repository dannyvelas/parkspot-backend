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
		apiRouter.Route("/admin", func(adminRouter chi.Router) {
			adminRouter.Use(jwtMiddleware.Authenticate) // jwtMiddleware.AuthenticateAdmin
			adminRouter.Get("/hello", sayHello())
			adminRouter.Get("/permits/active", getActive(permitRepo))
			adminRouter.Get("/permits", getAll(permitRepo))
			adminRouter.Get("/permits/exceptions", getExceptions(permitRepo))
			adminRouter.Get("/permits/expired", getExpired(permitRepo))
			adminRouter.Post("/permit", create(permitRepo, carRepo, residentRepo, dateFormat))
			adminRouter.Get("/residents", getAllResidents(residentRepo))
		})
		//apiRouter.Use(jwtMiddleware.AuthenticateUser)
		apiRouter.Post("/permit", create(permitRepo, carRepo, residentRepo, dateFormat))
	})

	return
}
