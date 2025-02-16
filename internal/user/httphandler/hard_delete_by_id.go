package httphandler

import (
	"fmt"
	"net/http"

	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/jkitajima/responder"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	OperationHardDeleteByID = "hard_delete_by_id"
	FileHardDeleteByID      = OperationHardDeleteByID + ".go"
)

func (s *UserServer) handleUserHardDeleteByID() http.HandlerFunc {
	const self = "handleUserHardDeleteByID"

	type request struct {
		Password string `json:"password" validate:"required"`
	}

	contract := map[string]responder.Field{
		"Password": {
			Name:       "password",
			Validation: "Field value cannot be an empty string.",
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)

		_, claims, err := jwtauth.FromContext(ctx)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Bearer token is malformatted.")
			return
		}

		sub, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Invalid UUID.")
			return
		}

		id := r.PathValue("userID")
		uuid, err := uuid.Parse(id)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "User ID must be a valid UUID.")
			return
		}

		if sub != uuid {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusForbidden, "You are not allowed to request deletion of another user.")
			return
		}

		req, err := responder.Decode[request](r)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Request body is invalid.")
			return
		}

		if errors := responder.ValidateInput(s.inputValidator, req, contract); len(errors) > 0 {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			responder.RespondClientErrors(w, r, errors...)
			return
		}

		err = s.service.HardDeleteByID(ctx, user.HardDeleteByIDRequest{ID: uuid, Password: req.Password})
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			switch err {
			case user.ErrNotFoundByID:
				responder.RespondMetaMessage(w, r, http.StatusNotFound, "Could not find any user with provided ID.")
			case user.ErrInvalidCredentials:
				responder.RespondMetaMessage(w, r, http.StatusBadRequest, "Provided credentials was invalid.")
			default:
				responder.RespondInternalError(w, r)
			}
			return
		}

		if err := responder.Respond(w, r, http.StatusNoContent, nil); err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("%s failed", OperationHardDeleteByID))
			span.RecordError(err)
			s.logger.ErrorContext(ctx, otel.FormatLog(Path, FileHardDeleteByID, self, "failed to encode response", err))
			responder.RespondInternalError(w, r)
			return
		}
	}

	otelhandler := otelhttp.NewHandler(http.HandlerFunc(handler), OperationHardDeleteByID)
	return otelhandler.ServeHTTP
}
