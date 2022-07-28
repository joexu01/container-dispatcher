package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"gorm.io/gorm"
)

type ContainerUser struct {
	UserId      int    `json:"user_id" gorm:"column:user_id"`
	ContainerId string `json:"container_id" gorm:"column:container_id"`
}

func (c *ContainerUser) TableName() string {
	return "container_user"
}

//func (c *ContainerUser) RetrieveUserContainerIDs()  {
//
//}

func (c *ContainerUser) GetContainerList(_ *gin.Context, tx *gorm.DB, param *dto.UserContainerListQueryInput) (info []dto.UserContainerInfo, err error) {
	query := tx.Raw("SELECT `user`.`id`, `user`.`username`, `container_user`.`container_id` FROM `user` JOIN `container_user` ON `user`.`id` = `container_user`.`user_id`")
	offset := (param.PageNo - 1) * param.PageSize

	if err := query.Limit(param.PageSize).Offset(offset).Order("`user`.`id` desc").Find(&info).Error; err != gorm.ErrRecordNotFound && err != nil {
		return nil, err
	}
	//query.Limit(param.PageSize).Offset(offset).Count(&total)
	return info, nil
}
