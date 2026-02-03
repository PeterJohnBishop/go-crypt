package server

import (
	"database/sql"
	"go-crypt/server/auth"
	"go-crypt/server/sqldb"
	"go-crypt/server/websockets"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var db *sql.DB

func ServeGin() {

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	// WEBSOCKET
	hub := websockets.NewHub()
	go hub.Run()

	// POSTGRES
	var sqlService sqldb.SqlService
	db := sqlService.ConnectPSQL()
	if db == nil {
		log.Fatal("Error: Unable to connect to Postgres.")
	}

	sqldb.CreateUsersTable(db)

	// SERVER
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// ROUTES
	protected := r.Group("/api")
	protected.Use(auth.JWTMiddleware())
	addOpenRoutes(r, db)
	addProtectedRoutes(protected, db)

	r.GET("/ws", func(c *gin.Context) {
		websockets.ServeWs(hub, c)
	})

	log.Printf("Serving Gin at :%s", appPort)
	r.Run(":" + appPort)
}
