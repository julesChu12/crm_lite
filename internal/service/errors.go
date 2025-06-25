package service

import "errors"

// 公共的业务错误定义
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrRoleNotFound      = errors.New("role not found")
)
