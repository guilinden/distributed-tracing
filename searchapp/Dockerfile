#Build stage

FROM golang:1.21.4-alpine3.18 AS build-stage

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /searchapp

#Deploy stage

FROM alpine:3.18

WORKDIR /

COPY --from=build-stage /searchapp /searchapp

EXPOSE 8081

RUN adduser -D searchapp

USER searchapp:searchapp

ENTRYPOINT ["/searchapp"]