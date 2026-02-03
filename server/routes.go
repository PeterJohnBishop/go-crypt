package server

import (
	"database/sql"
	"go-crypt/server/sqldb"

	"github.com/gin-gonic/gin"
)

func addOpenRoutes(r *gin.Engine, db *sql.DB) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "go-svr",
		})
	})
	r.POST("auth/login", func(c *gin.Context) {
		sqldb.Login(db, c)
	})
	r.POST("auth/register", func(c *gin.Context) {
		sqldb.RegisterUser(db, c)
	})
	r.GET("auth/refresh", func(c *gin.Context) {
		sqldb.Refresh(c)
	})
	r.POST("auth/logout", func(c *gin.Context) {
		sqldb.Logout(c)
	})
}

func addProtectedRoutes(r *gin.RouterGroup, db *sql.DB) {

	r.GET("/users", func(c *gin.Context) {
		sqldb.GetUsers(db, c)
	})
	r.GET("/users/:id", func(c *gin.Context) {
		sqldb.GetUserByID(db, c)
	})
	r.PUT("/users", func(c *gin.Context) {
		sqldb.UpdateUser(db, c)
	})
	r.DELETE("/users/:id", func(c *gin.Context) {
		sqldb.DeleteUserByID(db, c)
	})
}
