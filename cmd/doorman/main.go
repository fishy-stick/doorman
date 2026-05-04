package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"doorman/internal/auth"
	"doorman/internal/config"
	"doorman/internal/handler"
	"doorman/internal/store"
	"doorman/internal/webui"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s, err := store.New(cfg.Server.DB)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer s.Close()

	_, err = s.InitAdmin()
	if err != nil {
		log.Fatalf("Failed to initialize admin: %v", err)
	}

	sm := auth.NewSessionManager()

	r := gin.Default()

	knockHandler := handler.NewKnockHandler(s, cfg.Server.TrustProxyEnabled())
	adminHandler := handler.NewAdminHandler(s, sm)

	r.GET("/knock", auth.KnockAuth(s), knockHandler.Handle)

	admin := r.Group("/admin/api")
	{
		admin.POST("/login", adminHandler.Login)
		admin.POST("/logout", adminHandler.Logout)

		protected := admin.Group("")
		protected.Use(auth.AdminAuth(sm))
		{
			protected.GET("/session", adminHandler.Session)
			protected.PUT("/password", adminHandler.ChangePassword)
			protected.GET("/networks", adminHandler.ListNetworks)
			protected.GET("/networks/:id", adminHandler.GetNetwork)
			protected.POST("/networks", adminHandler.CreateNetwork)
			protected.PUT("/networks/:id", adminHandler.UpdateNetwork)
			protected.POST("/networks/:id/token", adminHandler.RegenerateNetworkToken)
			protected.DELETE("/networks/:id", adminHandler.DeleteNetwork)
			protected.GET("/networks/:id/knocks", adminHandler.ListKnocks)
		}
	}

	webui.Register(r)

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
