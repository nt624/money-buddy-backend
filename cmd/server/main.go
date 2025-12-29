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
