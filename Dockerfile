FROM golang:tip-alpine3.22

COPY go.mod go.sum ./

RUN go mod download

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /ip_exporter

CMD ["/ip_exporter"]
