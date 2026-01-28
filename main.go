package main

import (
	"go-crypt/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	hub := websocket.NewHub()
	go hub.Run()
	r.GET("/ws/:email", func(c *gin.Context) {
		websocket.ServeWs(hub, c)
	})
	r.Run(":8080")
}
