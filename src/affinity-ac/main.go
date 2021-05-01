package main

import (
	"fmt"
	"os"
	"os/signal"
	"flag"
	"crypto/tls"
	"net/http"
	"syscall"
	"context"
)

type myServerHandler struct {

}

var (
	tlscert , tlskey string
)


func getEnv(key , fallback string)  string {

	value, err := os.LookupEnv(key)

	if !err {
		value = fallback
	}
	return value
}

func main() {
	fmt.Println("Starting Admission controller image")

	certpem := getEnv("CERT_FILE", "/etc/certs/cert.pem")
	keypem := getEnv("KEY_FILE", "/etc/certs/key.pem")
	port := getEnv("PORT", "8443")

	flag.StringVar(&tlscert , "tlsCertFile", certpem , "the File Contains the Server Certificate for HTTPS")
	flag.StringVar(&tlskey, "tlsKeyFile", keypem , "The File Contains the Server key for HTTPS")

	flag.Parse()

	certs , err := tls.LoadX509KeyPair(tlscert, tlskey)

	if err != nil {
		fmt.Fprintf(os.Stderr , "Fail to Load Certificate/Key Pair: %v\n", err)
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%v", port),
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{certs}},
	}

	vac := myServerHandler{}
	mac := myServerHandler{}
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", mac.mutserve)
	mux.HandleFunc("/validate", vac.valserve)
	server.Handler = mux


	go func() {
		if err := server.ListenAndServeTLS("",""); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to Listen and Serve Web Hook Server Error: %v", err);
		}
	}()

	fmt.Fprintf(os.Stdout, "the Server is running on port : %s \n", port)


	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	fmt.Fprintf(os.Stdout , "Shutdown signal received , Shutting down the webhook gracefully...\n")
	server.Shutdown(context.Background())
}