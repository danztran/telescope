FROM docker.chotot.org/golang-builder:1.13.7-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
ENV GOPRIVATE "git.chotot.org/*"
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main

FROM alpine:3.11
WORKDIR /app
RUN apk update
RUN apk add --update ca-certificates
RUN apk add --no-cache tzdata && apk add --no-cache gcc musl-dev && \
  cp -f /usr/share/zoneinfo/Asia/Ho_Chi_Minh /etc/localtime && \
  apk del tzdata
COPY --from=builder /app/main .
COPY --from=builder /app/config .
CMD ["./main"]
