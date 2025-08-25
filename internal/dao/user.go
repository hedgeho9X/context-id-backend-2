package dao

import (
	"context"
	"context-id-backend/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

type UserDao struct{}

var User = &UserDao{}

// GetByUsername 根据用户名获取用户
func (d *UserDao) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user *model.User
	err := g.DB().Model("users").Where("username", username).Scan(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetByEmail 根据邮箱获取用户
func (d *UserDao) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
	err := g.DB().Model("users").Where("email", email).Scan(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Create 创建用户
func (d *UserDao) Create(ctx context.Context, user *model.User) error {
	_, err := g.DB().Model("users").Data(user).Insert()
	return err
}

// Update 更新用户
func (d *UserDao) Update(ctx context.Context, user *model.User) error {
	_, err := g.DB().Model("users").Data(user).Where("id", user.Id).Update()
	return err
}

// GetById 根据ID获取用户
func (d *UserDao) GetById(ctx context.Context, id uint64) (*model.User, error) {
	var user *model.User
	err := g.DB().Model("users").Where("id", id).Scan(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}
