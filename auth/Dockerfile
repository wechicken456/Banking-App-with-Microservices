FROM golang:1.24.4-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 

COPY . .

# CGO_ENABLED=0: Disables CGO to build a statically linked binary,
RUN CGO_ENABLED=0 go build -o /app/dist/main ./main.go 






FROM alpine:latest

WORKDIR /app

COPY --from=build /app/dist/main /app/main

RUN chmod +x /app/main

CMD ["/app/main"]
