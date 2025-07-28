package main

import (
	"log"
	"net/http"

	spotifyclient "github.com/ngomez18/playlist-router/internal/clients/spotify"
	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/controllers"
	"github.com/ngomez18/playlist-router/internal/middleware"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/pb"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/ngomez18/playlist-router/internal/static"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type AppDependencies struct {
	config       *config.Config
	repositories Repositories
	services     Services
	controllers  Controllers
	middleware   Middleware
}

type Repositories struct {
	basePlaylistRepository       repositories.BasePlaylistRepository
	userRepository               repositories.UserRepository
	spotifyIntegrationRepository repositories.SpotifyIntegrationRepository
}

type Services struct {
	basePlaylistService       services.BasePlaylistServicer
	userService               services.UserServicer
	spotifyIntegrationService services.SpotifyIntegrationServicer
	authService               services.AuthServicer
}

type Controllers struct {
	basePlaylistController controllers.BasePlaylistController
	authController         controllers.AuthController
}

type Middleware struct {
	auth *middleware.AuthMiddleware
}

func main() {
	var deps AppDependencies
	app := pocketbase.New()

	// Setup routes
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		if err := e.Next(); err != nil {
			return err
		}

		deps = initAppDependencies(app)

		if err := initCollections(app); err != nil {
			return err
		}

		return nil
	})

	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		initAppRoutes(deps, e)
		return e.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func initAppDependencies(app *pocketbase.PocketBase) AppDependencies {
	logger := app.Logger()
	cfg := config.MustLoad()

	spotifyClient := spotifyclient.NewSpotifyClient(&cfg.Auth, logger)

	repositories := Repositories{
		basePlaylistRepository:       pb.NewBasePlaylistRepositoryPocketbase(app),
		userRepository:               pb.NewUserRepositoryPocketbase(app),
		spotifyIntegrationRepository: pb.NewSpotifyIntegrationRepositoryPocketbase(app),
	}

	userService := services.NewUserService(repositories.userRepository, logger)
	spotifyIntegrationService := services.NewSpotifyIntegrationService(repositories.spotifyIntegrationRepository, logger)

	serviceInstances := Services{
		basePlaylistService:       services.NewBasePlaylistService(repositories.basePlaylistRepository, logger),
		userService:               userService,
		spotifyIntegrationService: spotifyIntegrationService,
		authService:               services.NewAuthService(userService, spotifyIntegrationService, spotifyClient, logger),
	}

	controllers := Controllers{
		basePlaylistController: *controllers.NewBasePlaylistController(serviceInstances.basePlaylistService),
		authController:         *controllers.NewAuthController(serviceInstances.authService, cfg),
	}

	middleware := Middleware{
		auth: middleware.NewAuthMiddleware(userService),
	}

	return AppDependencies{
		config:       cfg,
		repositories: repositories,
		services:     serviceInstances,
		controllers:  controllers,
		middleware:   middleware,
	}
}

func initAppRoutes(deps AppDependencies, e *core.ServeEvent) {
	// Auth routes (public - no middleware)
	auth := e.Router.Group("/auth")
	auth.GET("/spotify/login", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyLogin)))
	auth.GET("/spotify/callback", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyCallback)))

	// Protected API routes (require authentication)
	api := e.Router.Group("/api")
	
	// Auth validation endpoint (protected - uses middleware for validation)
	apiAuth := api.Group("/auth")
	apiAuth.GET("/validate", apis.WrapStdHandler(deps.middleware.auth.RequireAuth(http.HandlerFunc(deps.controllers.authController.ValidateToken))))
	
	// Base Playlist routes (protected)
	basePlaylist := api.Group("/base_playlist")
	basePlaylist.POST("", apis.WrapStdHandler(deps.middleware.auth.RequireAuth(http.HandlerFunc(deps.controllers.basePlaylistController.Create))))
	basePlaylist.GET("/{id}", apis.WrapStdHandler(deps.middleware.auth.RequireAuth(http.HandlerFunc(deps.controllers.basePlaylistController.GetByID))))
	basePlaylist.DELETE("/{id}", apis.WrapStdHandler(deps.middleware.auth.RequireAuth(http.HandlerFunc(deps.controllers.basePlaylistController.Delete))))

	// Serve static files (must be after API routes)
	setupStaticFileServer(e)
}

func setupStaticFileServer(e *core.ServeEvent) {
	fsys, err := static.GetFrontendFS()
	if err != nil {
		log.Fatal(err)
	}

	e.Router.GET("/", apis.Static(fsys, false))
}
