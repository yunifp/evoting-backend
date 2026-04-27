package main

import (
	"evoting-backend/internal/config"
	"evoting-backend/internal/handlers"
	"evoting-backend/internal/middlewares"
	"evoting-backend/internal/seeders"
	"log"
	"os"
	"time" 

	"github.com/gin-contrib/cors" 

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()
	seeders.RunSeeder(config.DB)
	
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders: []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))
	
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong! Server Golang siap untuk E-Voting.",
		})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		menuGroup := api.Group("/menus")
		menuGroup.Use(middlewares.RequireAuth())
		{
			menuGroup.GET("/me", handlers.GetMyMenus)
		}

		
		clientGroup := api.Group("/client")
		clientGroup.Use(middlewares.RequireAuth())
		{
			clientGroup.GET("/layanan", handlers.GetLayanan) 
			clientGroup.POST("/transactions", handlers.CreateTransaction)
			clientGroup.GET("/transactions/me", handlers.GetMyTransactions)
			clientGroup.GET("/pemilu", handlers.GetMyPemilu)
			clientGroup.POST("/pemilu", handlers.CreatePemilu)
			clientGroup.POST("/pemilu/:pemiluId/kandidat", handlers.AddKandidat) 
			clientGroup.DELETE("/kandidat/:id", handlers.DeleteKandidat) 
			clientGroup.POST("/pemilu/:pemiluId/dpt", handlers.AddDPT)
			clientGroup.GET("/pemilu/:pemiluId/dpt", handlers.GetDPTByPemilu)
		}

		admin := api.Group("/admin")
		admin.Use(middlewares.RequireAuth())
		{
			admin.PUT("/clients/:id/approve",
				middlewares.RequirePermission("Manajemen Client", "approve"),
				handlers.ApproveClient,
			)

			roles := admin.Group("/roles")
			{
				roles.GET("", middlewares.RequirePermission("Manajemen Role", "read"), handlers.GetRoles)
				roles.POST("", middlewares.RequirePermission("Manajemen Role", "create"), handlers.CreateRole)
				roles.PUT("/:id", middlewares.RequirePermission("Manajemen Role", "update"), handlers.UpdateRole)
				roles.DELETE("/:id", middlewares.RequirePermission("Manajemen Role", "delete"), handlers.DeleteRole)
				roles.POST("/:id/permissions", middlewares.RequirePermission("Manajemen Role", "assign_permission"), handlers.AssignPermissions)
			}

			menus := admin.Group("/menus")
			{
				menus.GET("", middlewares.RequirePermission("Manajemen Menu", "read"), handlers.GetAllMenus)
				menus.POST("", middlewares.RequirePermission("Manajemen Menu", "create"), handlers.CreateMenu)
				menus.PUT("/:id", middlewares.RequirePermission("Manajemen Menu", "update"), handlers.UpdateMenu)
				menus.DELETE("/:id", middlewares.RequirePermission("Manajemen Menu", "delete"), handlers.DeleteMenu)
			}

			permissions := admin.Group("/permissions")
			{
				permissions.GET("", middlewares.RequirePermission("Manajemen Hak Akses", "read"), handlers.GetPermissions)
				permissions.POST("", middlewares.RequirePermission("Manajemen Hak Akses", "create"), handlers.CreatePermission)
				permissions.PUT("/:id", middlewares.RequirePermission("Manajemen Hak Akses", "update"), handlers.UpdatePermission)
				permissions.DELETE("/:id", middlewares.RequirePermission("Manajemen Hak Akses", "delete"), handlers.DeletePermission)
			}

			layanan := admin.Group("/layanan")
			{
				layanan.GET("", middlewares.RequirePermission("Manajemen Layanan", "read"), handlers.GetLayanan)
				layanan.POST("", middlewares.RequirePermission("Manajemen Layanan", "create"), handlers.CreateLayanan)
				layanan.PUT("/:id", middlewares.RequirePermission("Manajemen Layanan", "update"), handlers.UpdateLayanan)
				layanan.DELETE("/:id", middlewares.RequirePermission("Manajemen Layanan", "delete"), handlers.DeleteLayanan)
			}

			transactions := admin.Group("/transactions")
			{
				transactions.GET("", middlewares.RequirePermission("Manajemen Transaksi", "read"), handlers.GetAllTransactions)
				transactions.PUT("/:id/approve", middlewares.RequirePermission("Manajemen Transaksi", "approve"), handlers.ApproveTransaction)
			}
		}
	}
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	router.Run(":" + port)
}