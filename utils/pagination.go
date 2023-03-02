package utils

import (
	"gin-bookstore/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PaginateScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		sort := c.DefaultQuery("sort", "id")
		offset := page * pageSize
		return db.Order(sort).Offset(offset).Limit(pageSize)
	}
}

func Paginate(c *gin.Context, model any, res any) error {
	var count int64
	if err := models.DB.Model(model).Count(&count).Scopes(PaginateScope(c)).Find(res).Error; err != nil {
		return err
	}
	c.Header("X-Total", strconv.FormatInt(count, 10))
	return nil
}
