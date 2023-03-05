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

func FilterScope(c *gin.Context, model any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		filter := c.QueryMap("filter")
		fieldNames := getFieldNames("filter", model)
		if len(filter) == 0 || len(fieldNames) == 0 {
			return db
		}
		for field, value := range filter {
			addFilter(db, fieldNames, field, value)
		}
		return db
	}
}

func addFilter(db *gorm.DB, fieldNames []string, field, value string) {
	split := strings.Split(field, ":")
	if slices.Contains(fieldNames, split[0]) {
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
}

func toAnyList[T any](input []T) []any {
	list := make([]any, len(input))
	for i, v := range input {
		list[i] = v
	}
	return list
}

func SearchScope(c *gin.Context, model any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		search := c.Query("search")
		fieldNames := getFieldNames("search", model)
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

func getFieldNames(name string, model any) []string {
	fieldNames := []string{}
	t := reflect.TypeOf(model).Elem()
	for i := 0; i < t.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := t.Field(i)

		// Get the field tag value
		tag := field.Tag.Get("pagination")
		if len(tag) > 0 {
			options := strings.Split(tag, ",")
			var fieldName string
			for _, option := range options {
				kv := strings.Split(option, "=")
				if strings.TrimSpace(kv[0]) != name {
					continue
				}
				if len(kv) == 1 {
					fieldName = strcase.ToSnake(field.Name)
				} else if len(kv) >= 2 {
					fieldName = strings.TrimSpace(kv[1])
				}
			}
			if len(fieldName) > 0 {
				fieldNames = append(fieldNames, fieldName)
			}
		}
	}
	return fieldNames
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

func Paginate(c *gin.Context, model any, res any) error {
	var count int64
	if err := models.DB.Model(model).Scopes(SearchScope(c, model), FilterScope(c, model)).Count(&count).Scopes(PaginateScope(c)).Find(res).Error; err != nil {
		return err
	}
	c.Header("X-Total", strconv.FormatInt(count, 10))
	return nil
}
