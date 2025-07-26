package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ngomez18/playlist-router/internal/clients"
	"github.com/ngomez18/playlist-router/internal/config"
	"github.com/ngomez18/playlist-router/internal/controllers"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/pb"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type AppDependencies struct {
	config       *config.Config
	repositories Repositories
	services     Services
	controllers  Controllers
}

type Repositories struct {
	basePlaylistRepository repositories.BasePlaylistRepository
}

type Services struct {
	basePlaylistService services.BasePlaylistServicer
	authService         services.AuthServicer
}

type Controllers struct {
	basePlaylistController controllers.BasePlaylistController
	authController         controllers.AuthController
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

	spotifyClient := clients.NewSpotifyClient(&cfg.Auth, logger)

	repositories := Repositories{
		basePlaylistRepository: pb.NewBasePlaylistRepositoryPocketbase(app),
	}

	services := Services{
		basePlaylistService: services.NewBasePlaylistService(repositories.basePlaylistRepository, logger),
		authService:         services.NewAuthService(app, spotifyClient, logger),
	}

	controllers := Controllers{
		basePlaylistController: *controllers.NewBasePlaylistController(services.basePlaylistService),
		authController:         *controllers.NewAuthController(services.authService),
	}

	return AppDependencies{
		config:       cfg,
		repositories: repositories,
		services:     services,
		controllers:  controllers,
	}
}

func initAppRoutes(deps AppDependencies, e *core.ServeEvent) {
	// Auth routes
	auth := e.Router.Group("/auth")
	auth.GET("/spotify/login", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyLogin)))
	auth.GET("/spotify/callback", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.authController.SpotifyCallback)))

	// Base Playlist routes
	basePlaylist := e.Router.Group("/api/base_playlist")
	basePlaylist.POST("", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.Create)))
	basePlaylist.GET("/{id}", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.GetByID)))
	basePlaylist.DELETE("/{id}", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.Delete)))

	// Serve static files last (catch-all)
	setupStaticFileServer(e)
}

func setupStaticFileServer(e *core.ServeEvent) {
	distPath := "web/dist"

	// Use PocketBase's static file serving for the entire frontend
	e.Router.GET("/*", apis.Static(os.DirFS(distPath), true))
}
