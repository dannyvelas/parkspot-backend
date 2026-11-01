# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/parkspot-backend .

FROM alpine:3.20
# this links the repo to the package. once a docker image with this label is pushed
# as a package, future pushes from that repo will succeed
# without this label, if the first push uses someone's PAT (e.g. if the first push is
# from someone's laptop instead of a github action) then future pushes will be
# restricted to that person
LABEL org.opencontainers.image.source=https://github.com/dannyvelas/parkspot-backend
RUN apk add --no-cache ca-certificates
COPY --from=build /out/parkspot-backend /usr/local/bin/parkspot-backend
EXPOSE 5000
ENTRYPOINT ["/usr/local/bin/parkspot-backend"]
