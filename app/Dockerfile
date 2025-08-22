
FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY go.mod ./
RUN go mod download


COPY . .


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o miapp .


FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app


COPY --from=builder --chown=nonroot:nonroot /app/miapp .

USER nonroot

EXPOSE 8080

CMD ["./miapp"]
