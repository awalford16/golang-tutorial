package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	protos "github.com/awalford16/golang-tutorial/currency/protos/currency"

	gorillahandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"

	"github.com/awalford16/golang-tutorial/products/data"
	"github.com/awalford16/golang-tutorial/products/handlers"
)

func main() {
	// env.Parse()

	l := hclog.Default()

	conn, err := grpc.Dial("localhost:9092", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create Protobuf client
	cc := protos.NewCurrencyClient(conn)

	// Create database instance
	db := data.NewProductsDB(cc, l)

	ph := handlers.NewProducts(l, db)
	sm := mux.NewRouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/", ph.GetProducts)
	getRouter.HandleFunc("/", ph.GetProducts).Queries("currency", "{[A-Z]{3}}")
	getRouter.HandleFunc("/{id:[0-9]+}", ph.GetProduct)
	getRouter.HandleFunc("/{id:[0-9]+}", ph.GetProduct).Queries("currency", "{[A-Z]{3}}")

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/{id:[0-9]+}", ph.UpdateProduct)
	putRouter.Use(ph.MiddlewareProductValidation)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/", ph.AddProduct)
	postRouter.Use(ph.MiddlewareProductValidation)

	// CORS
	ch := gorillahandlers.CORS(gorillahandlers.AllowedOrigins([]string{"http://localhost:3000"}))

	s := &http.Server{
		Addr:         ":9090",
		Handler:      ch(sm),
		ErrorLog:     l.StandardLogger(&hclog.StandardLoggerOptions{}),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	go func() {
		l.Info("Staarting server on port 9090")

		err := s.ListenAndServe()
		if err != nil {
			l.Error("Error starting server", err)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan
	l.Info("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
