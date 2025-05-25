FROM golang:1.23-alpine3.21 AS builder

LABEL authors="evan"
LABEL description="gatesession"
LABEL tags="gatesession"

WORKDIR /go/src/

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o /go/bin/gatesession app/gatesession/cmd/main.go


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /go/bin/gatesession /app/
COPY app/gatesession/etc/gatesession.yaml /etc/gatesession.yaml

EXPOSE 7900
EXPOSE 7901
EXPOSE 7905
EXPOSE 7910
EXPOSE 7915

CMD ["./gatesession", "-f", "/etc/gatesession.yaml"]