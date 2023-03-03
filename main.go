package main

import (
	"gin-bookstore/controllers"
	"gin-bookstore/models"

	_ "time/tzdata"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	models.ConnectDatabase()

	r.Static("/swagger-ui", "./swagger")

	books := r.Group("books")
	{
		books.GET("", controllers.FindBooks)
		books.POST("", controllers.CreateBook)
		books.GET(":id", controllers.FindBook)
		books.PATCH(":id", controllers.UpdateBook)
		books.DELETE(":id", controllers.DeleteBook)
	}

	r.Run()
}
