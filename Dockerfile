FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ai-code-tracker-server ./cmd/server

FROM alpine:3.21

RUN adduser -D -H tracker
USER tracker
COPY --from=build /out/ai-code-tracker-server /usr/local/bin/ai-code-tracker-server
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/ai-code-tracker-server"]
