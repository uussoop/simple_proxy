package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/uussoop/simple_proxy/api"
	"github.com/uussoop/simple_proxy/config"
	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/pkg/cache"
	mycron "github.com/uussoop/simple_proxy/pkg/cron"
	"github.com/uussoop/simple_proxy/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, dberr := gorm.Open(sqlite.Open("db/test.db"), &gorm.Config{})
	if dberr != nil {
		panic("failed to connect to database")
	}
	database.Db = db
	sqlDB, sqldberr := db.DB()
	if sqldberr != nil {
		panic("failed to get db")
	}

	migrationerr := db.AutoMigrate(&database.User{})
	defer sqlDB.Close()
	if migrationerr != nil {
		panic("failed to migrate")
	}
	mycron.Start()
	config.Init_users()

	c := cache.GetCache()
	allusr, err := database.GetAllUsers()
	if err != nil {
		return
	}
	for _, u := range allusr {
		c.Set(strconv.Itoa(int(u.ID))+"cachedusage", u.UsageToday, 0)
		c.Set(strconv.Itoa(int(u.ID)), u.UsageToday, 0)

	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	mux := http.NewServeMux()
	mux.HandleFunc("/", api.Forwarder)
	ctx, cancel := context.WithCancel(context.Background())

	server := &http.Server{
		Addr:        ":" + utils.Getenv("PORT", "8080"),
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
		// BaseContext: func(net.Listener) context.Context {

		// 	return context.WithValue(ctx, "config", config.Init_config())
		// },
	}

	go func() {
		<-sigCh
		fmt.Println("Received SIGINT signal")
		server.Shutdown(ctx)
		cancel() // Cancel the context when SIGINT is received
	}()
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
		return
	}
	if err != nil {
		fmt.Printf("error listening for server: %s\n", err)
	}
}
