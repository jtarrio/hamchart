# syntax=docker/dockerfile:1
FROM golang:1.21 AS build-stage

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch
WORKDIR /app
COPY --from=build-stage /build/hamchart ./
EXPOSE 8080
ENTRYPOINT [ "./hamchart", "-server_address=:8080" ]
