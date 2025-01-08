package httphandler

import (
	"fmt"
	"net/http"

	"auth/internal/auth"
	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"go.opentelemetry.io/otel/codes"
)

const FileRequestAccessToken = "request_access_token.go"

func (s *AuthServer) handleRequestAccessToken() http.HandlerFunc {
	const self = "handleRequestAccessToken"
	const tracename string = "auth_request_acess_token"

	// type request struct {
	// 	GrantType string
	// 	Username  string
	// 	Password  string
	// }

	type response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := s.tracer.Start(r.Context(), tracename)
		defer span.End()

		// Content-Type must be "application/x-www-form-urlencoded"
		if ctype := r.Header.Get("Content-Type"); ctype != "application/x-www-form-urlencoded" {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(fmt.Errorf("Content-Type was %s", ctype))
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Content-Type must be application/x-www-form-urlencoded")
			return
		}

		// Now that we know that the Content-Type is correct,
		// we validate the form values
		grantType := r.FormValue("grant_type")
		if grantType != "password" {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(fmt.Errorf("Content-Type was %s", grantType))
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "grant_type must be password")
			return
		}

		username := r.FormValue("username")
		if username == "" {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(fmt.Errorf("username was empty"))
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "username must not be empty")
			return
		}

		password := r.FormValue("password")
		if password == "" {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(fmt.Errorf("password was empty"))
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "password must not be empty")
			return
		}

		requestAcessTokenResponse, err := s.service.RequestAccessToken(ctx, auth.AccessTokenRequest{
			Username: username,
			Password: password,
		})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(err)
			switch err {
			case user.ErrNotFoundByEmail:
				responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Could not find any user with provided email.")
			case auth.ErrInvalidCredentials:
				responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Invalid credentials.")
			case user.ErrInternal:
				fallthrough
			case auth.ErrInternal:
				fallthrough
			default:
				responder.RespondInternalError(w, r)
			}
			return
		}

		resp := response{
			AccessToken: requestAcessTokenResponse.AccessToken,
			// Entity:    s.entity,
			// ID:        createResponse.User.ID,
			// Email:     createResponse.User.Email,
			// CreatedAt: createResponse.User.CreatedAt,
			// UpdatedAt: createResponse.User.UpdatedAt,
		}

		if err := responder.Respond(w, r, http.StatusCreated, &responder.DataField{Data: resp}); err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(err)
			s.logger.ErrorContext(ctx, otel.FormatLog(Path, FileRegisterUser, self, "failed to encode response", err))
			responder.RespondInternalError(w, r)
			return
		}
	}
}
