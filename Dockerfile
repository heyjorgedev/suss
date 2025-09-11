FROM node:21 AS frontend

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install

COPY . .
RUN npm run build

FROM golang:1.25-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend /app/http/dist/css /app/http/dist/css
RUN go tool templ generate
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build -o /app/suss cmd/suss/main.go

FROM alpine:latest
ENV PORT=8080
EXPOSE 8080

RUN apk add --no-cache sqlite-libs

COPY --from=builder /app/suss /app/suss

CMD ["/app/suss"]