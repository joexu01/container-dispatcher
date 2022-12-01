package dto

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/public"
)

type NewTaskParameters struct {
	Name          string `json:"name" gorm:"column:name" example:"测试任务"`
	Desc          string `json:"desc" gorm:"column:desc" example:"测试任务描述"`
	AlgorithmUuid string `json:"algorithm_uuid" gorm:"column:algorithm_uuid"`
}

func (param *NewTaskParameters) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

type TaskListQueryInput struct {
	PageNo   int `json:"page_no" form:"page_no" comment:"页数" example:"1" validate:"required"`        //页数
	PageSize int `json:"page_size" form:"page_size" comment:"每页条数" example:"20" validate:"required"` //每页条数
}

func (param *TaskListQueryInput) BindValidParam(c *gin.Context) error {
	return public.GetValidParamsDefault(c, param)
}

type DirFileInfo struct {
	Files []string `json:"files"`
	Count int      `json:"count"`
}
