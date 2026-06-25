FROM golang:1.20-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.work go.work.sum ./
COPY currency/go.mod currency/go.sum ./currency/
COPY images-api/go.mod images-api/go.sum ./images-api/
COPY products-api/go.mod products-api/go.sum ./products-api/

WORKDIR /src/images-api
RUN go mod download

WORKDIR /src
COPY images-api ./images-api

WORKDIR /src/images-api
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/images-api .

FROM scratch

COPY --from=build /out/images-api /images-api

EXPOSE 9091

ENTRYPOINT ["/images-api"]
