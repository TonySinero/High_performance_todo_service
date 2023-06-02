package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(port string, router *gin.Engine) error {
	server := &http.Server{
		Addr:         ":" + "8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      router,
		ErrorLog:     log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error running server: %s", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

//Code for HTTPS server
//in main.go
// serverCertFile := os.Getenv("serverCertFilePath")
// serverKeyFile := os.Getenv("serverKeyFilePath")
// go func() {
// 	if err := server.Run(port, routes, serverCertFile, serverKeyFile); err != nil {
// 		logrus.Fatalf("Error occurred while running HTTPS server: %s", err.Error())
// 	}
// }()

//in file server.go
// func (s *Server) Run(port string, router *gin.Engine, serverCert, serverKey string) error {
// 	// Create a CA certificate pool and add cert.pem to it
// 	caCert, err := os.ReadFile(serverCert)
// 	if err != nil {
// 		return fmt.Errorf("failed to read CA certificate: %s", err)
// 	}
// 	caCertPool := x509.NewCertPool()
// 	caCertPool.AppendCertsFromPEM(caCert)

// 	// Create the TLS Config with the CA pool and enable Client certificate validation
// 	tlsConfig := &tls.Config{
// 		ClientCAs:  caCertPool,
// 		ClientAuth: tls.RequireAndVerifyClientCert,
// 	}

// 	// Create a Server instance to listen on the specified port with the TLS config
// 	server := &http.Server{
// 		Addr:         ":" + port,
// 		ReadTimeout:  10 * time.Second,
// 		WriteTimeout: 10 * time.Second,
// 		Handler:      router,
// 		TLSConfig:    tlsConfig,
// 		ErrorLog:     log.New(os.Stderr, "ERROR: ", log.LstdFlags),
// 	}

// 	// Listen to HTTPS connections with the server certificate and key
// 	err = server.ListenAndServeTLS(serverCert, serverKey)
// 	if err != nil {
// 		return fmt.Errorf("error running server: %s", err)
// 	}
// 	return nil
// }

// func (s *Server) Shutdown(ctx context.Context) error {
// 	return s.httpServer.Shutdown(ctx)
// }
