package main

import "net/http"

func (app *app) routes() http.Handler {
	app.Router.HandlerFunc(http.MethodDelete, "/v1/users", app.deleteUserHandler)
	app.Router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	app.Router.HandlerFunc(http.MethodPut, "/v1/users/activate", app.activateUserHandler)

	app.Router.HandlerFunc(http.MethodPost, "/v1/users/authenticate", app.authUserHandler)

	app.Router.HandlerFunc(http.MethodPost, "/v1/tokens", app.createAuthTokenHandler)

	return app.Metrics(app.RecoverPanic(app.Router))
}
