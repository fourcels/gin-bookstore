package utils

import (
	"fmt"
	"gin-bookstore/models"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func FilterScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		filter := c.QueryMap("filter")
		if len(filter) == 0 {
			return db
		}
		for field, value := range filter {
			addFilter(db, field, value)
		}
		return db
	}
}

func addFilter(db *gorm.DB, field, value string) {
	split := strings.Split(field, ":")
	if len(split) == 1 {
		db.Where(fmt.Sprintf("%s = ?", split[0]), value)
	} else if len(split) >= 2 {
		switch split[1] {
		case "eq":
			db.Where(fmt.Sprintf("%s = ?", split[0]), value)
		case "ne":
			db.Where(fmt.Sprintf("%s <> ?", split[0]), value)
		case "lt":
			db.Where(fmt.Sprintf("%s < ?", split[0]), value)
		case "le":
			db.Where(fmt.Sprintf("%s <= ?", split[0]), value)
		case "gt":
			db.Where(fmt.Sprintf("%s > ?", split[0]), value)
		case "ge":
			db.Where(fmt.Sprintf("%s >= ?", split[0]), value)
		case "between":
			db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", split[0]), toAnyList(strings.Split(value, ","))...)
		case "contains", "like":
			db.Where(fmt.Sprintf("%s LIKE ?", split[0]), "%"+value+"%")
		case "startsWith":
			db.Where(fmt.Sprintf("%s = ?", split[0]), "%"+value)
		case "in":
			db.Where(fmt.Sprintf("%s IN ?", split[0]), strings.Split(value, ","))
		}
	}
}

func toAnyList[T any](input []T) []any {
	list := make([]any, len(input))
	for i, v := range input {
		list[i] = v
	}
	return list
}

func SearchScope(c *gin.Context, searchFields []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		search := c.Query("search")
		if len(searchFields) == 0 || len(search) == 0 {
			return db
		}
		db2 := db.Session(&gorm.Session{})
		for _, field := range searchFields {
			db2 = db2.Or(fmt.Sprintf("%s LIKE ?", field), "%"+search+"%")
		}
		return db.Where(db2)
	}
}

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

func Paginate(c *gin.Context, model any, res any, searchFields ...string) error {
	var count int64
	if err := models.DB.Model(model).Scopes(SearchScope(c, searchFields), FilterScope((c))).Count(&count).Scopes(PaginateScope(c)).Find(res).Error; err != nil {
		return err
	}
	c.Header("X-Total", strconv.FormatInt(count, 10))
	return nil
}
