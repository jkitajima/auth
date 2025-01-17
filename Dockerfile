# syntax=docker/dockerfile:1
FROM golang:1.23-alpine3.21 AS build
WORKDIR /repo
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o ./bin/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /bin
USER nonroot:nonroot
EXPOSE 80 443
COPY --from=build /repo/env.local.yaml ./
COPY --from=build /repo/bin/server ./
ENTRYPOINT [ "/bin/server" ]