package httphandler

import (
	"fmt"
	"net/http"

	"auth/internal/auth"
	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	OperationRequestAccessToken = "request_access_token"
	FileRequestAccessToken      = OperationRequestAccessToken + ".go"
)

func (s *AuthServer) handleRequestAccessToken() http.HandlerFunc {
	const self = "handleRequestAccessToken"

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

	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		req, err := decodeForm(r)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRequestAccessToken))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, err.Error())
			return
		}

		requestAcessTokenResponse, err := s.service.RequestAccessToken(ctx, auth.AccessTokenRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRequestAccessToken))
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

		s.tokensGeneratedCounter.Add(ctx, 1)

		resp := response{
			AccessToken: string(requestAcessTokenResponse.AccessToken),
			TokenType:   requestAcessTokenResponse.TokenType,
			ExpiresIn:   requestAcessTokenResponse.ExpiresIn,
		}

		if err := responder.Respond(w, r, http.StatusOK, resp); err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRequestAccessToken))
			span.RecordError(err)
			s.logger.ErrorContext(ctx, otel.FormatLog(Path, FileRegisterUser, self, "failed to encode response", err))
			responder.RespondInternalError(w, r)
			return
		}
	}

	otelhandler := otelhttp.NewHandler(http.HandlerFunc(handler), OperationRequestAccessToken)
	return otelhandler.ServeHTTP
}
