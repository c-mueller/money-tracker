FROM golang:1.25-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w \
      -X icekalt.dev/money-tracker/internal/buildinfo.Version=${VERSION} \
      -X icekalt.dev/money-tracker/internal/buildinfo.Commit=${COMMIT} \
      -X icekalt.dev/money-tracker/internal/buildinfo.BuildDate=${BUILD_DATE} \
      -X icekalt.dev/money-tracker/internal/buildinfo.GoVersion=$(go version | awk '{print $3}')" \
    -o /bin/money-tracker ./cmd/money-tracker

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/money-tracker /bin/money-tracker

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/bin/money-tracker"]
CMD ["serve"]
