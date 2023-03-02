package utils

import (
	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page int    `form:"page,default=0" binding:"min=0"`
	Size int    `form:"size,default=10" binding:"min=1"`
	Sort string `form:"sort,default=id"`
}

func BindPagination(c *gin.Context, pagination *Pagination) error {
	if err := c.ShouldBindQuery(pagination); err != nil {
		return err
	}
	return nil
}
