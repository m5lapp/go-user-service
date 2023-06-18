package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/m5lapp/go-user-service/internal/data"
	"github.com/m5lapp/go-user-service/serialisation/jsonz"
	"github.com/m5lapp/go-user-service/validator"
)

func (app *app) createAuthTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.InvalidCredentialsResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	if !match {
		app.InvalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}

	data := map[string]*data.Token{"authenticated_tokens": token}
	err = jsonz.WriteJSendSuccess(w, http.StatusCreated, nil, data)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
