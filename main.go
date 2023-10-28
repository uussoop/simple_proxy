package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	_ "github.com/GoAdminGroup/go-admin/adapter/gin"               // web framework adapter
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/sqlite" // sql driver
	_ "github.com/GoAdminGroup/themes/adminlte"

	"github.com/rodrikv/openai_proxy/api"
	"github.com/rodrikv/openai_proxy/internal/cron"
	"github.com/rodrikv/openai_proxy/middleware"
	"github.com/rodrikv/openai_proxy/middleware/auth"
	"github.com/rodrikv/openai_proxy/middleware/limit"
	"github.com/rodrikv/openai_proxy/panel"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	sigCh := make(chan os.Signal, 1)

	cron.Start()

	signal.Notify(sigCh, os.Interrupt)
	mux := http.NewServeMux()
	{
		mux.HandleFunc("/", api.Forwarder)
	}

	authMux := auth.IsAuthroized(limit.LimitToken(limit.LimitRequest(middleware.SetOpenAIServer(mux))))
	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:        ":" + utils.Getenv("PORT", "8080"),
		Handler:     authMux,
		BaseContext: func(net.Listener) context.Context { return ctx },
		// BaseContext: func(net.Listener) context.Context {

		// 	return context.WithValue(ctx, "config", config.Init_config())
		// },
	}

	p := panel.Panel{}
	go p.Run(utils.Getenv("PANEL_PORT", "8081"))

	go func() {
		<-sigCh
		fmt.Println("Received SIGINT signal")
		server.Shutdown(ctx)
		cancel() // Cancel the context when SIGINT is received
	}()
	fmt.Printf("listening for server: %s\n", server.Addr)
	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
		return
	}
	if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
