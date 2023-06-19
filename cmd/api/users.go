package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/m5lapp/go-user-service/internal/data"
	"github.com/m5lapp/go-service-toolkit/serialisation/jsonz"
	"github.com/m5lapp/go-service-toolkit/validator"
)

func (app *app) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email        string          `json:"email"`
		Password     string          `json:"password"`
		Name         string          `json:"name"`
		FriendlyName *string         `json:"friendly_name"`
		BirthDate    *jsonz.DateOnly `json:"birth_date,omitempty"`
		Gender       *string         `json:"gender,omitempty"`
		CountryCode  *string         `json:"country_code,omitempty"`
		TimeZone     *string         `json:"time_zone,omitempty"`
	}

	err := jsonz.ReadJSON(w, r, &input)

	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Email:        input.Email,
		Name:         input.Name,
		FriendlyName: input.FriendlyName,
		BirthDate:    input.BirthDate,
		Gender:       input.Gender,
		CountryCode:  input.CountryCode,
		TimeZone:     input.TimeZone,
		Activated:    false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.FailedValidationResponse(w, r, v.Errors)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	// Generate a token for the user to activate with.
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	// Send the user an activation email in the background.
	app.Background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.Logger.Error(err.Error())
		}
	})

	err = jsonz.WriteJSendSuccess(w, http.StatusAccepted, nil, jsonz.Envelope{"user": user})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *app) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateTokenPlaintext(v, input.TokenPlaintext)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.FailedValidationResponse(w, r, v.Errors)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.EditConflictResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
		return
	}

	err = jsonz.WriteJSendSuccess(w, http.StatusOK, nil, jsonz.Envelope{"user": user})
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}

func (app *app) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := jsonz.ReadJSON(w, r, &input)
	if err != nil {
		app.BadRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	if !v.Valid() {
		app.FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.DeleteByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = jsonz.WriteJSON(w, http.StatusNoContent, nil, nil)
	if err != nil {
		app.ServerErrorResponse(w, r, err)
	}
}
