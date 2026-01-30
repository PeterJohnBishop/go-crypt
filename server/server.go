package server

import (
	"fmt"
	"go-crypt/server/websockets"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ServeGin() {

	hub := websockets.NewHub()
	go hub.Run()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Database
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		host, user, password, dbname, dbPort,
	)

	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Waiting for DB... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		panic("failed to connect database")
	}
	if db != nil {
		log.Println("Connected to Postgres volume.")
	}

	r.GET("/ws", func(c *gin.Context) {
		websockets.ServeWs(hub, c)
	})

	log.Printf("Serving Gin at :%s", appPort)
	r.Run(":" + appPort)
}
