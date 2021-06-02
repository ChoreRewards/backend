package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	chorerewardsv1alpha1 "github.com/chorerewards/api/chorerewards/v1alpha1"
	"github.com/chorerewards/backend/server"
)

const httpEnabled = true

func main() {
	server := server.New()

	gServer := grpc.NewServer()

	chorerewardsv1alpha1.RegisterChoreRewardsServiceServer(gServer, server)
	reflection.Register(gServer)

	port := 8080
	addr := fmt.Sprintf(":%d", port)

	go httpProxyServer(8443, addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err, "Failed to create listener")
	}

	fmt.Println("Starting grpc server", "port", port)
	if err := gServer.Serve(listener); err != nil {
		log.Fatal(err, "Failed to start server")
	}
}

// httpProxyServer starts a new http server listening on the specified port, proxying
// requests to the provided grpc service
func httpProxyServer(port int, grpcAddr string) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if err := chorerewardsv1alpha1.RegisterChoreRewardsServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		log.Fatal(err, "Failed to register http handler")
	}

	// Create a handler for our multiplexer.
	h := Handler(mux)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		MaxAge:         int(time.Hour * 24),
	})

	h = c.Handler(h)

	log.Println("Starting http proxy server", "port", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), h), "Failed to start http proxy server")
}

func Handler(mux *runtime.ServeMux) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// checking Values as map[string][]string also catches ?pretty and ?pretty=
			// r.URL.Query().Get("pretty") would not.
			if _, ok := r.URL.Query()["pretty"]; ok {
				r.Header.Set("Accept", "application/json+pretty")
			}

			h.ServeHTTP(w, r)
		})
	}(mux)
}
