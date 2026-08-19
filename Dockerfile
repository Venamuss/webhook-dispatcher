# FROM golang:1.26.5-alpine AS builder-base

# WORKDIR /app

# COPY go.mod go.sum ./
# RUN --network=host go mod download

# Build the binaries for the different services
FROM golang:1.26.5-alpine AS build
WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOD=linux go build -o /bin/migrator ./cmd/migrator/main.go
RUN CGO_ENABLED=0 GOOD=linux go build -o /bin/dispatcher ./cmd/dispatcher/main.go
RUN CGO_ENABLED=0 GOOD=linux go build -o /bin/subscription ./cmd/subscription/main.go
RUN CGO_ENABLED=0 GOOD=linux go build -o /bin/gateway ./cmd/gateway/main.go

# Create a minimal image for the migrator binary
FROM alpine:3.19 AS migrator
WORKDIR /app
COPY --from=build /bin/migrator /bin/migrator
CMD ["/bin/migrator"]

# Create a minimal image for the dispatcher binary
FROM alpine:3.19 AS dispatcher
WORKDIR /app
COPY --from=build /bin/dispatcher /bin/dispatcher
CMD ["/bin/dispatcher"]

# Create a minimal image for the subscription binary
FROM alpine:3.19 AS subscription
WORKDIR /app
COPY --from=build /bin/subscription /bin/subscription
EXPOSE 8080
CMD ["/bin/subscription"]

# Create a minimal image for the gateway binary
FROM alpine:3.19 AS gateway
WORKDIR /app
COPY --from=build /bin/gateway /bin/gateway
EXPOSE 50051 8081
CMD ["/bin/gateway"]