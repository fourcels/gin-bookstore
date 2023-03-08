package main

import (
	"embed"
	"gin-bookstore/controllers"
	"gin-bookstore/models"
	"io/fs"
	"net/http"

	_ "time/tzdata"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
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
	models.ConnectDatabase()

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// r.Static("/swagger-ui", "./swagger")
	r.StaticFS("/swagger-ui", SwaggerFS())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	r.Use(cors.New(corsConfig))
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
