package main

import (
	dbgen "money-buddy-backend/db/generated"
	"money-buddy-backend/infra/repository"
	"money-buddy-backend/internal/db"
	"money-buddy-backend/internal/handlers"
	"money-buddy-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
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
	repo := repository.NewExpenseRepositorySQLC(queries)
	categoryRepo := repository.NewCategoryRepositorySQLC(queries)
	service := services.NewExpenseService(repo, categoryRepo)
	handlers.NewExpenseHandler(r, service)

	categoryService := services.NewCategoryService(categoryRepo)
	handlers.NewCategoryHandler(r, categoryService)

	userRepo := repository.NewUserRepositorySQLC(queries)
	fixedCostRepo := repository.NewFixedCostRepositorySQLC(queries)
	txManager := db.NewSQLTxManager(dbConn)
	initialSetupService := services.NewInitialSetupService(userRepo, fixedCostRepo, txManager)
	handlers.NewInitialSetupHandler(r, initialSetupService)

	r.Run() // デフォルトで:8080で起動
}
