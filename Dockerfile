FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o webhook .



FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/webhook .
EXPOSE 8443
CMD ["./webhook"]
