package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/auth"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/eventengine"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/admin"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/cart"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/inventory"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/product"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/session"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/features/user"
	"github.com/eng-by-sjb/yellow-pines-e-commerce-backend/internal/middlewares"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"golang.org/x/sync/errgroup"
)

type ServerConfig struct {
	Addr         string
	DB           *sql.DB
	TokenManager *auth.TokenService
}

type server struct {
	*ServerConfig

	doneCh        chan struct{}   // used to signal internal go routines to shutdown
	internalSrvWG *sync.WaitGroup // used to wait for all internal go routines within individual routes to finish before shutting down the server.

	eventEngine eventengine.SubscribeRegisterPublisher
	srv         *http.Server
}

func NewServer(serverConfig *ServerConfig) *server {
	srv := &server{
		ServerConfig:  serverConfig,
		doneCh:        make(chan struct{}),
		internalSrvWG: &sync.WaitGroup{},
	}

	return srv
}

func (s *server) Run() {
	router := chi.NewRouter()

	// strip trailing slashes at the end of the url
	// e.g. /users/1/ -> /users/1
	// this middleware should be applied to all routes
	// to ensure that the url is correctly formatted
	router.Use(chimiddleware.StripSlashes)

	s.prep()

	router.Mount("/api/v1", s.v1Router()) // api version 1 subrouter

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.Addr),
		Handler: router,
	}

	// start server and listen for [os.Signal] signals to graceful shutdown server.
	s.listenAndServe()
}

func (s *server) listenAndServe() {
	shutdownCtx, shutdownCancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer shutdownCancel()

	errGrp, shutdownCtx := errgroup.WithContext(shutdownCtx)

	errGrp.Go(
		func() error {
			log.Printf("server started and is listening at port %s...\n", s.Addr)

			if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) && err != nil {
				return fmt.Errorf("failed to start server: %w", err)
			}

			return nil
		},
	)

	errGrp.Go(
		func() error {
			<-shutdownCtx.Done() // block and listen shutdown signals
			println()
			log.Println("hold and wait, server is gracefully shutting down...")

			ctx, cancel := context.WithTimeout(
				context.Background(),
				(20 * time.Second),
			)
			defer cancel()

			log.Println("server has stopped receiving new requests")
			log.Println("waiting for all pending requests to finish....")
			if err := s.srv.Shutdown(ctx); err != nil {
				return fmt.Errorf("server failed shutdown gracefully: %w", err)
			}

			return nil
		},
	)

	if err := errGrp.Wait(); err != nil {
		log.Fatal(err.Error())
	}
	log.Println("all pending requests completed!")

	log.Println("waiting for all internal pending go routines....")
	close(s.doneCh)
	s.internalSrvWG.Wait()
	log.Println("all internal go routines are done")

	log.Println("closing other resources...")
	if err := s.DB.Close(); err != nil {
		log.Println("server failed to close db for shutdown")
	}

	log.Println("server has been gracefully shutdown")
	os.Exit(0)
}

// prep prepares server dependencies needed for server to function
func (s *server) prep() {
	s.eventEngine = eventengine.NewEventEngine(
		&eventengine.EventEngineConfig{
			DoneCh:        s.doneCh,
			InternalSrvWG: s.internalSrvWG,
		},
	)
}

func (s *server) v1Router() *chi.Mux {
	r := chi.NewRouter()

	// health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Println("health check")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// todo: pass the [s.internalSrvWG] wait group down to the Various handlers, services and so on that spawn internal go routines.

	// session feature
	sessionStore := session.NewStore(s.DB)
	sessionService := session.NewService(
		sessionStore,
		s.TokenManager,
	)
	sessionHandler := session.NewHandler(sessionService)
	sessionHandler.RegisterRoutes(r)

	// user feature
	userStore := user.NewStore(s.DB)
	userService := user.NewService(userStore, sessionService)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(r)

	//admin feature
	adminStore := admin.NewStore(s.DB)
	adminService := admin.NewService(
		adminStore,
		sessionService,
	)
	adminHandler := admin.NewHandler(adminService)
	adminHandler.RegisterRoutes(r)

	//middleware
	middleware := middlewares.NewMiddleware(
		s.TokenManager,
	)

	// inventory feature
	inventoryStore := inventory.NewStore(s.DB)
	inventoryService := inventory.NewService(
		inventoryStore,
	)

	// products feature
	productStore := product.NewStore(s.DB)
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
