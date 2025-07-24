FROM golang:1.24.3 AS builder
WORKDIR /app
COPY . .
RUN go build -o app ./main

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/config/config.yaml ./config/config.yaml
EXPOSE 8080
CMD ["./app"]