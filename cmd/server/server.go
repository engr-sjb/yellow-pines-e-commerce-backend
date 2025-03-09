package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/admin"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/cart"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/inventory"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/product"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/session"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/user"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/middlewares"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
)

type Server struct {
	addr         string
	db           *sql.DB
	tokenService *auth.TokenService
}

func NewServer(addr string, db *sql.DB, tokenService *auth.TokenService) *Server {
	return &Server{
		addr:         addr,
		db:           db,
		tokenService: tokenService,
	}
}

func (s *Server) Start() error {
	router := chi.NewRouter()

	// strip trailing slashes at the end of the url
	// e.g. /users/1/ -> /users/1
	// this middleware should be applied to all routes
	// to ensure that the url is correctly formatted
	router.Use(chimiddleware.StripSlashes)

	router.Mount("/api/v1", s.v1Router()) // api version 1 subrouter

	srv := http.Server{
		Addr:    fmt.Sprintf(":%s", s.addr),
		Handler: router,
	}

	log.Printf("Server started at port %s\n", s.addr)

	return srv.ListenAndServe()
}

func (s *Server) v1Router() *chi.Mux {
	r := chi.NewRouter()

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("health check")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// session feature
	sessionStore := session.NewStore(s.db)
	sessionService := session.NewService(
		sessionStore,
		s.tokenService,
	)
	sessionHandler := session.NewHandler(sessionService)
	sessionHandler.RegisterRoutes(r)

	// user feature
	userStore := user.NewStore(s.db)
	userService := user.NewService(userStore, sessionService)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(r)

	//admin feature
	adminStore := admin.NewStore(s.db)
	adminService := admin.NewService(
		adminStore,
		sessionService,
	)
	adminHandler := admin.NewHandler(adminService)
	adminHandler.RegisterRoutes(r)

	//middleware
	middleware := middlewares.NewMiddleware(
		s.tokenService,
	)

	// inventory feature
	inventoryStore := inventory.NewStore(s.db)
	inventoryService := inventory.NewService(
		inventoryStore,
	)

	// products feature
	productStore := product.NewStore(s.db)
	productService := product.NewService(
		productStore,
		inventoryService,
	)
	productHandler := product.NewHandler(
		productService,
		middleware,
	)
	productHandler.RegisterRoutes(r)

	return r
}
