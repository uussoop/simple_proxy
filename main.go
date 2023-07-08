package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/uussoop/simple_proxy/api"
	"github.com/uussoop/simple_proxy/config"
	"github.com/uussoop/simple_proxy/utils"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.Forwarder)

	ctx := context.Background()
	server := &http.Server{
		Addr:    ":" + utils.Getenv("PORT", "8080"),
		Handler: mux,
		BaseContext: func(net.Listener) context.Context {

			return context.WithValue(ctx, "config", config.Init_config())
		},
	}

	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
		return
	}
	if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
