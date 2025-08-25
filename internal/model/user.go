package model

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// User 用户模型
type User struct {
	Id          uint64      `json:"id" db:"id"`
	Username    string      `json:"username" db:"username"`
	Email       string      `json:"email" db:"email"`
	DisplayName string      `json:"displayName" db:"display_name"`
	Avatar      string      `json:"avatar" db:"avatar"`
	Phone       string      `json:"phone" db:"phone"`
	Status      int         `json:"status" db:"status"`
	CreatedAt   *gtime.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   *gtime.Time `json:"updatedAt" db:"updated_at"`
}

// UserLoginReq 用户登录请求
type UserLoginReq struct {
	Code  string `json:"code" v:"required#授权码不能为空"`
	State string `json:"state" v:"required#状态码不能为空"`
}

// UserLoginRes 用户登录响应
type UserLoginRes struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// UserInfoRes 用户信息响应
type UserInfoRes struct {
	User *User `json:"user"`
}
