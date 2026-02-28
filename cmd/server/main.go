package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"dispatchpro/internal/config"
	"dispatchpro/internal/handlers"
	"dispatchpro/internal/middlewares"
	"dispatchpro/internal/repositories"
	"dispatchpro/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	wd, _ := os.Getwd()

	if err := godotenv.Load(filepath.Join(wd, ".env")); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	if _, err := config.ConnectDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repositories.NewProductRepository().CreateIndexes(ctx)
	repositories.NewOrderRepository().CreateIndexes(ctx)
	repositories.NewDriverRepository().CreateIndexes(ctx)
	repositories.NewUserRepository().CreateIndexes(ctx)

	authService := services.NewAuthService()
	productService := services.NewProductService()
	orderService := services.NewOrderService()
	driverService := services.NewDriverService()

	authHandler := handlers.NewAuthHandler(authService, cfg)
	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)
	driverHandler := handlers.NewDriverHandler(driverService)

	authMiddleware := middlewares.NewAuthMiddleware(cfg)
	rateLimiter := middlewares.NewRateLimiter(cfg.RateLimitReq, cfg.RateLimitWindow)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middlewares.CORSMiddleware())
	r.Use(middlewares.ErrorHandler())

	r.LoadHTMLGlob(filepath.Join(wd, "web/templates/*"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "DispatchPro - Sistema Logístico",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login - DispatchPro",
		})
	})

	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title": "Registro - DispatchPro",
		})
	})

	api := r.Group("/api")
	api.Use(middlewares.RateLimit(rateLimiter))
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			users := protected.Group("/users")
			{
				users.GET("/profile", authHandler.GetProfile)
				users.GET("", authHandler.GetUsers)
			}

			products := protected.Group("/products")
			{
				products.POST("", productHandler.CreateProduct)
				products.GET("", productHandler.GetProducts)
				products.GET("/low-stock", productHandler.GetLowStock)
				products.GET("/:id", productHandler.GetProduct)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.POST("/:id/stock", productHandler.AdjustStock)
			}

			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetOrders)
				orders.GET("/stats", orderHandler.GetStats)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
				orders.POST("/:id/assign", orderHandler.AssignDriver)
			}

			drivers := protected.Group("/drivers")
			{
				drivers.POST("", driverHandler.CreateDriver)
				drivers.GET("", driverHandler.GetDrivers)
				drivers.GET("/:id", driverHandler.GetDriver)
				drivers.PUT("/:id", driverHandler.UpdateDriver)
			}
		}
	}

	web := r.Group("/dashboard")
	web.Use(authMiddleware.RequireAuth())
	{
		web.GET("", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"title": "Dashboard - DispatchPro",
			})
		})

		web.GET("/orders", func(c *gin.Context) {
			c.HTML(http.StatusOK, "orders.html", gin.H{
				"title": "Pedidos - DispatchPro",
			})
		})

		web.GET("/products", func(c *gin.Context) {
			c.HTML(http.StatusOK, "products.html", gin.H{
				"title": "Inventario - DispatchPro",
			})
		})

		web.GET("/drivers", func(c *gin.Context) {
			c.HTML(http.StatusOK, "drivers.html", gin.H{
				"title": "Repartidores - DispatchPro",
			})
		})
	}

	r.Static("/static", filepath.Join(wd, "web/static"))

	port := cfg.Port
	log.Printf("Server starting on port %s", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
