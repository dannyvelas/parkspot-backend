FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/parkspot-backend ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/parkspot-backend /usr/local/bin/parkspot-backend
COPY --from=build /out/migrate /usr/local/bin/migrate
EXPOSE 5000
ENTRYPOINT ["/usr/local/bin/parkspot-backend"]
