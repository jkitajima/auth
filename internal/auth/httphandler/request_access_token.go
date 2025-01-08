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

	type request struct {
		GrantType string
		Username  string
		Password  string
	}

	type response struct {
		AccessToken string `json:"access_token"`
		// RefreshToken string `json:"refresh_token"`
		TokenType string `json:"token_type"`
		ExpiresIn int    `json:"expires_in"`
	}

	decodeForm := func(r *http.Request) (request, error) {
		// Content-Type must be "application/x-www-form-urlencoded"
		if ctype := r.Header.Get("Content-Type"); ctype != "application/x-www-form-urlencoded" {
			return request{}, fmt.Errorf("Content-Type must be application/x-www-form-urlencoded")
		}

		// Now that we know that the Content-Type is correct,
		// we validate the form values
		grantType := r.FormValue("grant_type")
		if grantType != "password" {
			return request{}, fmt.Errorf("grant_type must be password")
		}

		username := r.FormValue("username")
		if username == "" {
			return request{}, fmt.Errorf("username must not be empty")
		}

		password := r.FormValue("password")
		if password == "" {
			return request{}, fmt.Errorf("password must not be empty")
		}

		return request{
			GrantType: grantType,
			Username:  username,
			Password:  password,
		}, nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := s.tracer.Start(r.Context(), tracename)
		defer span.End()

		req, err := decodeForm(r)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, err.Error())
			return
		}

		requestAcessTokenResponse, err := s.service.RequestAccessToken(ctx, auth.AccessTokenRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(err)
			switch err {
			case user.ErrNotFoundByEmail:
				fallthrough
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
			AccessToken: string(requestAcessTokenResponse.AccessToken),
			TokenType:   requestAcessTokenResponse.TokenType,
			ExpiresIn:   requestAcessTokenResponse.ExpiresIn,
		}

		if err := responder.Respond(w, r, http.StatusOK, resp); err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", tracename))
			span.RecordError(err)
			s.logger.ErrorContext(ctx, otel.FormatLog(Path, FileRegisterUser, self, "failed to encode response", err))
			responder.RespondInternalError(w, r)
			return
		}
	}
}
