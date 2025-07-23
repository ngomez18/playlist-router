package main

import (
	"log"
	"net/http"

	"github.com/ngomez18/playlist-router/internal/controllers"
	"github.com/ngomez18/playlist-router/internal/repositories"
	"github.com/ngomez18/playlist-router/internal/repositories/pb"
	"github.com/ngomez18/playlist-router/internal/services"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

type AppDependencies struct {
    repositories Repositories
    services Services
    controllers Controllers
}

type Repositories struct {
    basePlaylistRepository repositories.BasePlaylistRepository
}

type Services struct {
    basePlaylistService services.BasePlaylistServicer
}

type Controllers struct {
    basePlaylistController controllers.BasePlaylistController
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
    repositories := Repositories{
        basePlaylistRepository: pb.NewBasePlaylistRepositoryPocketbase(app),
    }

    services := Services{
        basePlaylistService: services.NewBasePlaylistService(repositories.basePlaylistRepository),
    }

    controllers := Controllers{
        basePlaylistController: *controllers.NewBasePlaylistController(services.basePlaylistService),
    }

    return AppDependencies{
        repositories: repositories,
        services: services,
        controllers: controllers,
    }
}


func initAppRoutes(deps AppDependencies, e *core.ServeEvent) {
    // Base Playlist routes
    basePlaylist := e.Router.Group("/api/base_playlist")
    basePlaylist.POST("", apis.WrapStdHandler(http.HandlerFunc(deps.controllers.basePlaylistController.Create)))
}
