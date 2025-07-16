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
	ErrUserRoleNotFound      = errors.New("user role assignment not found")
	ErrPermissionNotFound    = errors.New("permission not found")
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrForbidden             = errors.New("operation not permitted")
	ErrInsufficientBalance   = errors.New("insufficient balance")
)
