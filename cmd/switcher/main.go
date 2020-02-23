package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	caFile       = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	tokenFile    = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	enablePatch  = "[{\"op\": \"add\", \"path\": \"/metadata/annotations/fencing~1enabled\", \"value\": \"true\"}]"
	disablePatch = "[{\"op\": \"remove\", \"path\": \"/metadata/annotations/fencing~1enabled\"}]"
)

func main() {

	node := loadEnv("NODE_NAME")
	host := loadEnv("KUBERNETES_SERVICE_HOST")
	port := loadEnv("KUBERNETES_PORT_443_TCP_PORT")

	// Load CA cert
	caCert := loadFile(caFile)
	token := string(loadFile(tokenFile))

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}
	apply := func(data string) {
		req, err := http.NewRequest(
			"PATCH", "https://"+host+":"+port+"/api/v1/nodes/"+node, bytes.NewBuffer([]byte(data)),
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Content-Type", "application/json-patch+json")
		_, err = client.Do(req)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("enable fencing for", node)
	apply(enablePatch)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		fmt.Println("disable fencing for", node)
		apply(disablePatch)
		done <- true
	}()

	<-done
}

func loadEnv(e string) string {
	v := os.Getenv(e)
	if v == "" {
		fmt.Println("env " + e + " is not specfied.")
		os.Exit(1)
	}
	return v
}

func loadFile(f string) []byte {
	v, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return v
}
