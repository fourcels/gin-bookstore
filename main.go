package main

import (
	"embed"
	"gin-bookstore/controllers"
	"gin-bookstore/models"
	"io/fs"
	"net/http"

	_ "time/tzdata"

	"github.com/gin-gonic/gin"
)

//go:embed swagger
var swagger embed.FS

func SwaggerFS() http.FileSystem {
	sub, err := fs.Sub(swagger, "swagger")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

func main() {
	r := gin.Default()

	models.ConnectDatabase()

	// r.Static("/swagger-ui", "./swagger")
	r.StaticFS("/swagger-ui", SwaggerFS())

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
