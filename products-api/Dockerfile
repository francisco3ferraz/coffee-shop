FROM golang:1.20-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.work go.work.sum ./
COPY currency/go.mod currency/go.sum ./currency/
COPY images-api/go.mod images-api/go.sum ./images-api/
COPY products-api/go.mod products-api/go.sum ./products-api/

WORKDIR /src/products-api
RUN go mod download

WORKDIR /src
COPY currency ./currency
COPY products-api ./products-api

WORKDIR /src/products-api
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/products-api .

FROM scratch

WORKDIR /

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/products-api /products-api
COPY products-api/swagger.yaml /swagger.yaml

EXPOSE 9090

USER 65532:65532

ENTRYPOINT ["/products-api"]
