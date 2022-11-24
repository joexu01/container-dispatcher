package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/public"
	"time"
)

type UserInfoOutput struct {
	Id           int       `json:"id"`
	Username     string    `json:"username"`
	LoginTime    time.Time `json:"login_time"`
	Avatar       string    `json:"avatar"`
	Introduction string    `json:"introduction"`
	Roles        []string  `json:"roles"`
}

type ChangePwdInput struct {
	OldPassword string `json:"old_password" form:"old_password" comment:"旧密码" example:"123456" validate:"required"` // 旧密码
	NewPassword string `json:"new_password" form:"new_password" comment:"新密码" example:"123456" validate:"required"` // 新密码
}

func (param *ChangePwdInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

type UserLoginInput struct {
	Username string `json:"username" form:"username" comment:"用户名" example:"joexu01" validate:"required,validate_username"` // 管理员用户名
	Password string `json:"password" form:"password" comment:"密码" example:"12345678" validate:"required"`                   // 管理员密码
}

type UserSessionInfo struct {
	Id        int       `json:"id"`
	UserName  string    `json:"username"`
	LoginTime time.Time `json:"login_time"`
}

func (param *UserLoginInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

/*
id: undefined,
        username: undefined,
        email: undefined,
        introduction: undefined,
        createdAt: undefined,
        role: ''
*/

type UserLoginOutput struct {
	Id        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	Role      string    `json:"role"`
}

type NewUserInput struct {
	Username    string `json:"username" validate:"required"`
	RawPassword string `json:"raw_password" validate:"required"`
	Email       string `json:"email" validate:"required"`
	UserRole    int    `json:"user_role" gorm:"column:user_role" validate:"required"`
}

func (param *NewUserInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

type User struct {
	Username string `json:"username" gorm:"column:username" validate:"required"`
	Password string `json:"password" gorm:"column:hashed_password" validate:"required"`
	Email    string `json:"email" gorm:"column:email" validate:"required"`
}

type UserListQueryInput struct {
	PageNo   int `json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required"`        //页数
	PageSize int `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` //每页条数
}

func (param *UserListQueryInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}
