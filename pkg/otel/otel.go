package otel

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Route(router chi.Router, method string, pattern string, fn http.HandlerFunc) {
	handler := otelhttp.WithRouteTag(pattern, fn)
	switch method {
	case http.MethodGet:
		router.Get(pattern, handler.ServeHTTP)
	case http.MethodPost:
		router.Post(pattern, handler.ServeHTTP)
	case http.MethodPut:
		router.Put(pattern, handler.ServeHTTP)
	case http.MethodPatch:
		router.Patch(pattern, handler.ServeHTTP)
	case http.MethodDelete:
		router.Delete(pattern, handler.ServeHTTP)
	}
}

func FormatLog(prefix, filename, funcname, msg string, err error) string {
	format := "%s/%s: %s"
	if prefix == "" {
		format = format[3:]
	}

	if err == nil {
		format = format[:len(format)-4]
		if prefix == "" {
			return fmt.Sprintf(format, prefixMsg(filename, funcname, msg))
		}
		return fmt.Sprintf(format, prefix, prefixMsg(filename, funcname, msg))
	}

	return fmt.Sprintf(format, prefix, prefixMsg(filename, funcname, msg), err.Error())
}

func prefixMsg(filename, funcname, msg string) string {
	return fmt.Sprintf("%s [%s]: %s", filename, funcname, msg)
}
