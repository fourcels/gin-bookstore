package utils

import (
	"fmt"
	"gin-bookstore/models"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func FilterScope(c *gin.Context, fieldNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		filter := c.QueryMap("filter")
		if len(filter) == 0 || len(fieldNames) == 0 {
			return db
		}
		for field, value := range filter {
			arr := strings.Split(field, ":")
			field = strings.TrimSpace(arr[0])
			if slices.Contains(fieldNames, field) {
				var operator string
				if len(arr) > 1 {
					operator = strings.TrimSpace(arr[1])
				}
				addFilter(db, field, operator, strings.TrimSpace(value))
			}
		}
		return db
	}
}

func addFilter(db *gorm.DB, field, operator, value string) {
	switch operator {
	case "eq":
		db.Where(fmt.Sprintf("%s = ?", field), value)
	case "ne":
		db.Where(fmt.Sprintf("%s <> ?", field), value)
	case "lt":
		db.Where(fmt.Sprintf("%s < ?", field), value)
	case "le":
		db.Where(fmt.Sprintf("%s <= ?", field), value)
	case "gt":
		db.Where(fmt.Sprintf("%s > ?", field), value)
	case "ge":
		db.Where(fmt.Sprintf("%s >= ?", field), value)
	case "between":
		db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), toAnyList(strings.Split(value, ","))...)
	case "contains", "like":
		db.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	case "startsWith":
		db.Where(fmt.Sprintf("%s = ?", field), "%"+value)
	case "in":
		db.Where(fmt.Sprintf("%s IN ?", field), strings.Split(value, ","))
	default:
		db.Where(fmt.Sprintf("%s = ?", field), value)
	}
}

func toAnyList[T any](input []T) []any {
	list := make([]any, len(input))
	for i, v := range input {
		list[i] = v
	}
	return list
}

func SearchScope(c *gin.Context, fieldNames []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		search := strings.TrimSpace(c.Query("search"))
		if len(search) == 0 || len(fieldNames) == 0 {
			return db
		}

		db2 := db.Session(&gorm.Session{})
		for _, fieldName := range fieldNames {
			db2 = db2.Or(fmt.Sprintf("%s LIKE ?", fieldName), "%"+search+"%")
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

type PaginationConfig struct {
	Filter []string
	Search []string
}

func getConfig(model any) PaginationConfig {
	var config PaginationConfig
	t := reflect.TypeOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get("pagination")
		if len(tag) > 0 {
			items := strings.Split(tag, ",")
			for _, item := range items {
				arr := strings.Split(item, "=")
				var fieldName string
				if len(arr) > 1 {
					fieldName = strings.TrimSpace(arr[1])
				} else {
					fieldName = strcase.ToSnake(field.Name)
				}
				switch strings.TrimSpace(arr[0]) {
				case "filter":
					config.Filter = append(config.Filter, fieldName)
				case "search":
					config.Search = append(config.Search, fieldName)
				}
			}

		}
	}

	return config

}

func Paginate(c *gin.Context, model any, res any) error {
	var count int64
	config := getConfig(model)
	if err := models.DB.Model(model).Scopes(SearchScope(c, config.Search), FilterScope(c, config.Filter)).Count(&count).Scopes(PaginateScope(c)).Find(res).Error; err != nil {
		return err
	}
	c.Header("X-Total", strconv.FormatInt(count, 10))
	return nil
}
