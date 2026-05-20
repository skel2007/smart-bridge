FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/smart-bridge-server ./cmd/smart-bridge-server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/smart-bridge-server /smart-bridge-server

USER nonroot:nonroot

ENTRYPOINT ["/smart-bridge-server"]
CMD ["--config", "/secrets/smart-bridge/config.yaml"]
