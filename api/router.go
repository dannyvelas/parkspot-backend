package api

import (
	"github.com/dannyvelas/lasvistas_api/config"
	"github.com/dannyvelas/lasvistas_api/models"
	"github.com/dannyvelas/lasvistas_api/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"net/http"
	"strings"
)

func NewRouter(
	httpConfig config.HttpConfig,
	tokenConfig config.TokenConfig,
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
		MaxAge:           300,
	}))

	jwtMiddleware := NewJWTMiddleware(tokenConfig)

	// index
	router.Handle("/*", indexHandler())

	// api
	router.Route("/api", func(r chi.Router) {
		r.Group(func(anyoneRouter chi.Router) {
			anyoneRouter.Post("/login", login(jwtMiddleware, adminRepo, residentRepo))
			anyoneRouter.Post("/logout", logout())
		})

		r.Group(func(officeRouter chi.Router) {
			officeRouter.Use(jwtMiddleware.authenticate(AdminRole)) //, SecurityRole
			officeRouter.Get("/permits/active", getPermits(permitRepo, models.ActivePermits))
			officeRouter.Get("/permits", getPermits(permitRepo, models.AllPermits))
			officeRouter.Get("/permits/exceptions", getPermits(permitRepo, models.ExceptionPermits))
			officeRouter.Get("/permits/expired", getPermits(permitRepo, models.ExpiredPermits))
			officeRouter.Get("/residents", getAllResidents(residentRepo))
			officeRouter.Get("/resident/{id}", getOneResident(residentRepo))
			officeRouter.Get("/visitors", getAllVisitors(visitorRepo))
			officeRouter.Get("/visitors/search", searchVisitors(visitorRepo))
		})

		r.Group(func(userRouter chi.Router) {
			userRouter.Use(jwtMiddleware.authenticate(AdminRole, ResidentRole)) //, SecurityRole
			userRouter.Get("/me", getMe())
			userRouter.Get("/hello", sayHello())
			userRouter.Post("/permit", createPermit(permitRepo, residentRepo, carRepo, dateFormat))
			userRouter.Get("/permit/{id:[0-9]+}", getOnePermit(permitRepo))
			userRouter.Get("/car/{id}", getOneCar(carRepo))
			userRouter.Put("/car/{id}", editCar(carRepo))
			userRouter.Delete("/permit/{id:[0-9]+}", deletePermit(permitRepo, residentRepo, carRepo))
			userRouter.Get("/permits/search", searchPermits(permitRepo))
			userRouter.Get("/resident/{id}/permits", getAllPermitsOfResident(permitRepo))
			userRouter.Get("/resident/{id}/permits/active", getActivePermitsOfResident(permitRepo))
		})
	})

	return
}

func indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_app") {
			http.ServeFile(w, r, "client/build"+r.URL.Path)
			return
		}

		http.ServeFile(w, r, "client/build/index.html")
	}
}
