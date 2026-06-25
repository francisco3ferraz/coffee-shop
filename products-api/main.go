package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"

	protos "github.com/franciscofferraz/coffee-shop/currency/protos/currency"
	"github.com/franciscofferraz/coffee-shop/products-api/data"
	"github.com/franciscofferraz/coffee-shop/products-api/handlers"
	muxHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	l := hclog.Default()
	validation := data.NewValidation()

	currencyAddr := getEnv("CURRENCY_ADDR", "localhost:9092")
	bindAddr := getEnv("PRODUCTS_BIND_ADDR", "127.0.0.1:9090")
	conn, err := grpc.Dial(currencyAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	// create client
	cc := protos.NewCurrencyClient(conn)

	// create database instance
	db := data.NewProductsDB(cc, l)

	productsHandler := handlers.NewProducts(l, validation, db)

	serveMux := mux.NewRouter()

	// handlers for API
	getRequest := serveMux.Methods(http.MethodGet).Subrouter()
	getRequest.HandleFunc("/healthz", healthz)
	getRequest.HandleFunc("/products", productsHandler.ListAll).Queries("currency", "{[A-Z]{3}}")
	getRequest.HandleFunc("/products", productsHandler.ListAll)

	getRequest.HandleFunc("/products/{id:[0-9]+}", productsHandler.ListSingle).Queries("currency", "{[A-Z]{3}}")
	getRequest.HandleFunc("/products/{id:[0-9]+}", productsHandler.ListSingle)

	putRequest := serveMux.Methods(http.MethodPut).Subrouter()
	putRequest.HandleFunc("/products", productsHandler.Update)
	putRequest.Use(productsHandler.MiddlewareValidateProduct)

	postRequest := serveMux.Methods(http.MethodPost).Subrouter()
	postRequest.HandleFunc("/products", productsHandler.Create)
	postRequest.Use(productsHandler.MiddlewareValidateProduct)

	deleteRequest := serveMux.Methods(http.MethodDelete).Subrouter()
	deleteRequest.HandleFunc("/products/{id:[0-9]+}", productsHandler.Delete)

	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	swaggerHandler := middleware.Redoc(opts, nil)

	getRequest.Handle("/docs", swaggerHandler)
	getRequest.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	corsHandler := muxHandlers.CORS(muxHandlers.AllowedOrigins([]string{"*"}))

	// create a new server
	s := http.Server{
		Addr:         bindAddr,
		Handler:      corsHandler(serveMux),
		ErrorLog:     l.StandardLogger(&hclog.StandardLoggerOptions{}),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		l.Info("Starting server", "bind_address", bindAddr)

		err := s.ListenAndServe()
		if err != nil {
			l.Error("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	log.Println("Got signal:", sig)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func healthz(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
