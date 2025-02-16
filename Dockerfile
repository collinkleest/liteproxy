FROM golang:alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./src

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/main .

EXPOSE 8082

CMD ["/app/main"]