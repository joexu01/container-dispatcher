package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/public"
)

type UserContainerInfo struct {
	Id          int    `json:"id" gorm:"column:id"`
	Username    string `json:"username" gorm:"column:username"`
	ContainerId string `json:"container_id" gorm:"column:container_id"`
	Image       string `json:"image"`
	Command     string `json:"command"`
	Created     string `json:"created"`
	Status      string `json:"status"`
	Ports       string `json:"ports"`
	Name        string `json:"name"`
}

type UserContainerInfoList struct {
	List []*UserContainerInfo `json:"list"`
}

type MyContainerInfo struct {
	ContainerId string `json:"container_id" gorm:"column:container_id"`
	Image       string `json:"image"`
	Command     string `json:"command"`
	Created     string `json:"created"`
	Status      string `json:"status"`
	Ports       string `json:"ports"`
	Name        string `json:"name"`
}

type MyContainerInfoList struct {
	List []*MyContainerInfo `json:"list"`
}

type UserContainerListQueryInput struct {
	PageNo   int `json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required"`        //页数
	PageSize int `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` //每页条数
}

func (param *UserContainerListQueryInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}
