package main

import (
	"net"
	"os"

	"github.com/franciscofferraz/coffee-shop/currency/data"
	protos "github.com/franciscofferraz/coffee-shop/currency/protos/currency"
	"github.com/franciscofferraz/coffee-shop/currency/server"

	hclog "github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	log := hclog.Default()

	rates, err := data.NewRates(log)
	if err != nil {
		log.Error("Unable to generate rates", "error", err)
		os.Exit(1)
	}

	gs := grpc.NewServer()
	c := server.NewCurrency(rates, log)

	protos.RegisterCurrencyServer(gs, c)
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(gs, healthServer)

	reflection.Register(gs)

	bindAddr := getEnv("CURRENCY_BIND_ADDR", ":9092")
	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Error("Unable to listen", "error", err)
		os.Exit(1)
	}

	log.Info("Starting server", "bind_address", bindAddr)
	gs.Serve(l)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
