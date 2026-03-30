FROM golang:1.26.1-alpine AS build
WORKDIR /src
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /incident-mcp .

FROM gcr.io/distroless/static:nonroot
COPY --from=build /incident-mcp /incident-mcp
EXPOSE 8099
ENTRYPOINT ["/incident-mcp"]
