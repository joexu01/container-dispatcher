package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/public"
)

type AlgorithmParams struct {
	Name             string `json:"name" gorm:"column:name" example:"测试算法"`
	Desc             string `json:"desc" gorm:"column:desc" example:"测试算法描述"`
	EntryPoint       string `json:"entry_point" gorm:"column:entry_point" example:"start.py"`
	ExecBinary       string `json:"exec_binary" gorm:"column:exec_binary" example:"python3"`
	DefaultImageName string `json:"default_image_name" example:"full_env"`
}

func (param *AlgorithmParams) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

type AlgorithmListQueryInput struct {
	PageNo   int `json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required"`        //页数
	PageSize int `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` //每页条数
}

func (param *AlgorithmListQueryInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}
