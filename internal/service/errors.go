package service

import "errors"

// 公共的业务错误定义
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrInvalidToken          = errors.New("invalid or expired token")
	ErrRoleNotFound          = errors.New("role not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrPhoneAlreadyExists    = errors.New("phone number already exists")
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrRoleNameAlreadyExists = errors.New("role name already exists")
)
