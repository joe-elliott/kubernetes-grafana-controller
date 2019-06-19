FROM golang:1.12 as build
WORKDIR /src

COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  

WORKDIR /root/
COPY --from=build /src/app .

USER nobody:nobody

CMD ["./app"] 
