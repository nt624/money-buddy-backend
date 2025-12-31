package main

import (
	dbgen "money-buddy-backend/db/generated"
	"money-buddy-backend/internal/db"
	"money-buddy-backend/internal/handlers"
	"money-buddy-backend/internal/repositories"
	"money-buddy-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	dbConn, err := db.NewDB()
	if err != nil {
		panic(err)
	}

	queries := dbgen.New(dbConn)
	repo := repositories.NewExpenseRepositorySQLC(queries)
	service := services.NewExpenseService(repo)
	handlers.NewExpenseHandler(r, service)

	r.Run() // デフォルトで:8080で起動
}
