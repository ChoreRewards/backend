package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	chorerewardsv1alpha1 "github.com/chorerewards/api/chorerewards/v1alpha1"
	"github.com/chorerewards/backend/internal/auth"
	"github.com/chorerewards/backend/internal/server"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)

	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {
	// Config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.httpProxy.enabled", false)
	viper.SetDefault("server.httpProxy.port", 8443)

	// DB defaults
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.username", "chorerewards")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.name", "chorerewards")

	// Auth defaults
	viper.SetDefault("auth.key", "secretkey")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("unable to read config")
	}

	var (
		port             = viper.GetInt("server.port")
		httpProxyEnabled = viper.GetBool("server.httpProxy.enabled")
		httpProxyPort    = viper.GetInt("server.httpProxy.port")

		dbHost     = viper.GetString("db.host")
		dbPort     = viper.GetInt("db.port")
		dbUsername = viper.GetString("db.username")
		dbPassword = viper.GetString("db.password")
		dbName     = viper.GetString("db.name")

		authKey = viper.GetString("auth.key")
	)

	log.WithFields(log.Fields{
		"Server Port":        port,
		"HTTP Proxy Enabled": httpProxyEnabled,
		"HTTP Proxy Port":    httpProxyPort,
		"Database Name":      dbName,
		"Database Host":      dbHost,
		"Database Port":      dbPort,
		"Database Username":  dbUsername,
	}).Info("Config Initialised")

	tokenManager := auth.NewTokenManager(authKey)

	server, err := server.New(
		server.Config{DBHost: dbHost, DBPort: dbPort, DBUsername: dbUsername, DBPassword: dbPassword, DBName: dbName},
		tokenManager,
	)
	if err != nil {
		log.Fatalf("Unable to initialise new Server: %+v", err)
	}

	gServer := grpc.NewServer(
		grpc.UnaryInterceptor(tokenManager.ValidateAuthInterceptor),
	)

	chorerewardsv1alpha1.RegisterChoreRewardsServiceServer(gServer, server)

	reflection.Register(gServer)

	addr := fmt.Sprintf(":%d", port)

	if httpProxyEnabled {
		go httpProxyServer(httpProxyPort, addr)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err, "Failed to create listener")
	}

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting grpc server")

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

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting http proxy server")

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
