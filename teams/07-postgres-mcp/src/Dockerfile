FROM golang:1.26.1-alpine AS build
WORKDIR /src
COPY go.mod .
COPY main.go .
# go mod tidy downloads lib/pq and generates go.sum at build time.
RUN go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o /postgres-mcp .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /postgres-mcp /postgres-mcp
EXPOSE 8099
ENTRYPOINT ["/postgres-mcp"]
