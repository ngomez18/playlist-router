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
	childPlaylistRepository      repositories.ChildPlaylistRepository
	userRepository               repositories.UserRepository
	spotifyIntegrationRepository repositories.SpotifyIntegrationRepository
}

type Services struct {
	authService               services.AuthServicer
	userService               services.UserServicer
	basePlaylistService       services.BasePlaylistServicer
	childPlaylistService      services.ChildPlaylistServicer
	spotifyIntegrationService services.SpotifyIntegrationServicer
	spotifyApiService         services.SpotifyAPIServicer
}

type Controllers struct {
	basePlaylistController  controllers.BasePlaylistController
	childPlaylistController controllers.ChildPlaylistController
	authController          controllers.AuthController
	spotifyController       controllers.SpotifyController
}

type Middleware struct {
	auth        *middleware.AuthMiddleware
	spotifyAuth *middleware.SpotifyAuthMiddleware
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
		childPlaylistRepository:      pb.NewChildPlaylistRepositoryPocketbase(app),
		userRepository:               pb.NewUserRepositoryPocketbase(app),
		spotifyIntegrationRepository: pb.NewSpotifyIntegrationRepositoryPocketbase(app),
	}

	userService := services.NewUserService(repositories.userRepository, logger)
	spotifyIntegrationService := services.NewSpotifyIntegrationService(repositories.spotifyIntegrationRepository, logger)

	serviceInstances := Services{
		authService:               services.NewAuthService(userService, spotifyIntegrationService, spotifyClient, logger),
		userService:               userService,
		basePlaylistService:       services.NewBasePlaylistService(repositories.basePlaylistRepository, repositories.spotifyIntegrationRepository, spotifyClient, logger),
		childPlaylistService:      services.NewChildPlaylistService(repositories.childPlaylistRepository, repositories.basePlaylistRepository, repositories.spotifyIntegrationRepository, spotifyClient, logger),
		spotifyIntegrationService: spotifyIntegrationService,
		spotifyApiService:         services.NewSpotifyAPIService(spotifyClient, logger),
	}

	controllers := Controllers{
		basePlaylistController:  *controllers.NewBasePlaylistController(serviceInstances.basePlaylistService),
		childPlaylistController: *controllers.NewChildPlaylistController(serviceInstances.childPlaylistService),
		authController:          *controllers.NewAuthController(serviceInstances.authService, cfg),
		spotifyController:       *controllers.NewSpotifyController(serviceInstances.spotifyApiService),
	}

	middleware := Middleware{
		auth:        middleware.NewAuthMiddleware(userService),
		spotifyAuth: middleware.NewSpotifyAuthMiddleware(spotifyIntegrationService, spotifyClient, logger),
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
	// Auth routes
	auth := e.Router.Group("/auth")
	auth.GET("/spotify/login", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyLogin)))
	auth.GET("/spotify/callback", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyCallback)))
	auth.GET("/validate", apis.WrapStdHandler(deps.middleware.auth.RequireAuth(http.HandlerFunc(deps.controllers.authController.ValidateToken))))

	// Protected API routes (require authentication)
	api := e.Router.Group("/api")
	api.BindFunc(apis.WrapStdMiddleware(deps.middleware.auth.RequireAuth))

	// Base Playlist routes
	basePlaylist := api.Group("/base_playlist")
	basePlaylist.POST("", apis.WrapStdHandler(deps.middleware.spotifyAuth.RequireSpotifyAuth(http.HandlerFunc(deps.controllers.basePlaylistController.Create))))
	basePlaylist.GET("", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.GetByUserID)))
	basePlaylist.GET("/{id}", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.GetByID)))
	basePlaylist.DELETE("/{id}", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.Delete)))

	// Child Playlist routes for a specific base playlist
	basePlaylist.POST("/{basePlaylistID}/child_playlist", apis.WrapStdHandler(deps.middleware.spotifyAuth.RequireSpotifyAuth(http.HandlerFunc(deps.controllers.childPlaylistController.Create))))
	basePlaylist.GET("/{basePlaylistID}/child_playlist", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.childPlaylistController.GetByBasePlaylistID)))

	// Child Playlist routes by ID
	childPlaylist := api.Group("/child_playlist")
	childPlaylist.GET("/{id}", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.childPlaylistController.GetByID)))
	childPlaylist.PUT("/{id}", apis.WrapStdHandler(deps.middleware.spotifyAuth.RequireSpotifyAuth(http.HandlerFunc(deps.controllers.childPlaylistController.Update))))
	childPlaylist.DELETE("/{id}", apis.WrapStdHandler(deps.middleware.spotifyAuth.RequireSpotifyAuth(http.HandlerFunc(deps.controllers.childPlaylistController.Delete))))

	// Spotify routes (protected)
	spotify := api.Group("/spotify")
	spotify.BindFunc(apis.WrapStdMiddleware(deps.middleware.spotifyAuth.RequireSpotifyAuth))
	spotify.GET("/playlists", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.spotifyController.GetUserPlaylists)))

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
