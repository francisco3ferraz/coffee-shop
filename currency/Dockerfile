FROM golang:1.20-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.work go.work.sum ./
COPY currency/go.mod currency/go.sum ./currency/
COPY images-api/go.mod images-api/go.sum ./images-api/
COPY products-api/go.mod products-api/go.sum ./products-api/

WORKDIR /src/currency
RUN go mod download

WORKDIR /src
COPY currency ./currency

WORKDIR /src/currency
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/currency .

FROM scratch

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/currency /currency

EXPOSE 9092

ENTRYPOINT ["/currency"]
