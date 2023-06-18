package main

import (
	"embed"
	"flag"
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/m5lapp/go-user-service/config"
	"github.com/m5lapp/go-user-service/internal/data"
	"github.com/m5lapp/go-user-service/mailer"
	"github.com/m5lapp/go-user-service/persistence/sqldb"
	"github.com/m5lapp/go-user-service/vcs"
	"github.com/m5lapp/go-user-service/webapp"
	"golang.org/x/exp/slog"
)

//go:embed "templates"
var templateFS embed.FS

type appConfig struct {
	db   config.SqlDB
	smtp config.Smtp
}

type app struct {
	webapp.WebApp
	cfg    appConfig
	models data.Models
	mailer mailer.Mailer
}

func (app *app) routes() http.Handler {
	app.Router.HandlerFunc(http.MethodDelete, "/v1/users", app.deleteUserHandler)
	app.Router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	app.Router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)

	app.Router.HandlerFunc(http.MethodPost, "/v1/tokens", app.createAuthTokenHandler)

	return app.Metrics(app.RecoverPanic(app.Router))
}

func main() {
	var serverCfg config.Server
	var appCfg appConfig

	serverCfg.Flags(":8080")
	appCfg.db.Flags("postgres", 25, 25, "15m")
	appCfg.smtp.Flags("", "")

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", vcs.Version())
		os.Exit(0)
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	logger := slog.New(logHandler)
	db, err := sqldb.OpenDB(appCfg.db)
	if err != nil {
		logger.Error(err.Error(), nil)
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database connection pool established")

	app := &app{
		WebApp: webapp.New(serverCfg, logger),
		cfg:    appCfg,
		models: data.NewModels(db),
		mailer: mailer.New(&appCfg.smtp, templateFS),
	}

	err = app.Serve(app.routes())
	if err != nil {
		logger.Error(err.Error(), nil)
		os.Exit(1)
	}
}
