FROM golang:1.27-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLE=0 go build -o product-service .


FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/product-service .

EXPOSE 8080
CMD [ "./product-service" ]