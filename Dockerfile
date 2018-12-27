FROM golang:1.11 as build
WORKDIR /go/src/kubernetes-grafana-controller

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest  
WORKDIR /root/
COPY --from=build /go/src/kubernetes-grafana-controller/app .
CMD ["./app"] 