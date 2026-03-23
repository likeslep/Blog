// utils/db_errors.go
package utils

import (
	"github.com/go-sql-driver/mysql"
	"strings"
)

// 检查是否是重复条目错误
func IsDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}

	// MySQL错误
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return mysqlErr.Number == 1062 // 1062: Duplicate entry
	}

	// 通用错误信息
	return strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// 检查是否是外键约束错误
func IsForeignKeyError(err error) bool {
	if err == nil {
		return false
	}

	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		return mysqlErr.Number == 1451 || mysqlErr.Number == 1452 // 外键约束错误
	}

	return strings.Contains(err.Error(), "foreign key constraint")
}

// 包装数据库错误
func WrapDBError(err error, message string) error {
	if err == nil {
		return nil
	}

	if IsDuplicateEntryError(err) {
		return ErrUserAlreadyExists
	}

	if IsForeignKeyError(err) {
		return ErrBadRequest
	}

	if message == "" {
		message = "数据库操作失败"
	}

	return WrapError(ErrInternalServer, message)
}
