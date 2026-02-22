FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /bin/money-tracker ./cmd/money-tracker

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /bin/money-tracker /bin/money-tracker

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/bin/money-tracker"]
CMD ["serve"]
