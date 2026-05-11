FROM golang:latest AS build

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bankapi ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /bankapi /bankapi
COPY migrations ./migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/bankapi"]
