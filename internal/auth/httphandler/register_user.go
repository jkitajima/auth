package httphandler

import (
	"fmt"
	"net/http"
	"time"

	"auth/internal/auth"
	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	OperationRegisterUser = "register_user"
	FileRegisterUser      = OperationRegisterUser + ".go"
)

func (s *AuthServer) handleUserRegister() http.HandlerFunc {
	const self = "handleUserRegister"

	type request struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	type response struct {
		Entity    string    `json:"entity"`
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	contract := map[string]responder.Field{
		"Email": {
			Name:       "email",
			Validation: "Field is required and must be a valid email.",
		},
		"Password": {
			Name:       "password",
			Validation: "Field value cannot be an empty string.",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		req, err := responder.Decode[request](r)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRegisterUser))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Request body is invalid.")
			return
		}

		if errors := responder.ValidateInput(s.inputValidator, req, contract); len(errors) > 0 {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRegisterUser))
			span.RecordError(err)
			responder.RespondClientErrors(w, r, errors...)
			return
		}

		registerResponse, err := s.service.Register(ctx, auth.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRegisterUser))
			span.RecordError(err)
			switch err {
			case user.ErrEmailAlreadyInUse:
				responder.RespondMetaMessage(w, r, http.StatusConflict, "There is already an user with provided email.")
			case user.ErrInternal:
				fallthrough
			default:
				responder.RespondInternalError(w, r)
			}
			return
		}

		s.usersCreatedCounter.Add(ctx, 1)

		resp := response{
			Entity:    s.entity,
			ID:        registerResponse.User.ID,
			Email:     registerResponse.User.Email,
			CreatedAt: registerResponse.User.CreatedAt,
			UpdatedAt: registerResponse.User.UpdatedAt,
		}

		if err := responder.Respond(w, r, http.StatusCreated, &responder.DataField{Data: resp}); err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationRegisterUser))
			span.RecordError(err)
			s.logger.ErrorContext(ctx, otel.FormatLog(Path, FileRegisterUser, self, "failed to encode response", err))
			responder.RespondInternalError(w, r)
			return
		}
	}

	otelhandler := otelhttp.NewHandler(http.HandlerFunc(handler), OperationRegisterUser)
	return otelhandler.ServeHTTP
}
