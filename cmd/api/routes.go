package main

import "net/http"

func (app *app) routes() http.Handler {
	app.Router.HandlerFunc(http.MethodDelete, "/v1/user", app.deleteUserHandler)
	app.Router.HandlerFunc(http.MethodGet, "/v1/user/email/:value", app.getUserHandler)
	app.Router.HandlerFunc(http.MethodGet, "/v1/user/id/:value", app.getUserHandler)
	app.Router.HandlerFunc(http.MethodPost, "/v1/user", app.registerUserHandler)
	app.Router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.activateUserHandler)
	app.Router.HandlerFunc(http.MethodPost, "/v1/user/authenticate", app.authUserHandler)

	app.Router.HandlerFunc(http.MethodPost, "/v1/token", app.createAuthTokenHandler)

	return app.Metrics(app.RecoverPanic(app.Router))
}
