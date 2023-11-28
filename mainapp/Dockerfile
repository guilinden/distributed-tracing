#Build stage

FROM golang:1.21.4-alpine3.18 AS build-stage

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /mainapp

#Deploy stage

FROM alpine:3.18

WORKDIR /

COPY --from=build-stage /mainapp /mainapp

EXPOSE 8081

RUN adduser -D mainapp

USER mainapp:mainapp

ENTRYPOINT ["/mainapp"]