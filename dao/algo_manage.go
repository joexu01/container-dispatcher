package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"gorm.io/gorm"
	"time"
)

type Algorithm struct {
	//Id         int    `json:"id" gorm:"column:id"`\
	Uuid       string    `json:"uuid" gorm:"column:uuid;primary_key"`
	Name       string    `json:"name" gorm:"column:name"`
	Desc       string    `json:"desc" gorm:"column:desc"`
	Path       string    `json:"path" gorm:"column:path"`
	EntryPoint string    `json:"entry_point" gorm:"column:entry_point"`
	ExecBinary string    `json:"exec_binary" gorm:"column:exec_binary"`
	AuthorId   int       `json:"author_id" gorm:"column:author_id"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	Files      string    `json:"files" gorm:"column:files"`
}

func (a *Algorithm) TableName() string {
	return "algorithm"
}

func (a *Algorithm) Find(_ *gin.Context, tx *gorm.DB, search *Algorithm) (*Algorithm, error) {
	out := &Algorithm{}
	err := tx.Model(out).Where(search).Find(out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (a *Algorithm) PageList(_ *gin.Context, tx *gorm.DB, param *dto.AlgorithmListQueryInput) (total int64, algoList []AlgorithmResult, err error) {
	query := tx.Table(a.TableName()).Joins("join user on user.id = algorithm.author_id")
	// .Where("user.is_delete=0")

	offset := (param.PageNo - 1) * param.PageSize
	//Limit(param.PageSize).Offset(offset)
	if err = query.Limit(param.PageSize).Offset(offset).Order("algorithm.created_at desc").Scan(&algoList).Error; err != gorm.ErrRecordNotFound && err != nil {
		return 0, nil, err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return total, algoList, nil
}

func (a *Algorithm) Create(_ *gin.Context, tx *gorm.DB) error {
	return tx.Create(a).Error
}

func (a *Algorithm) Update(_ *gin.Context, tx *gorm.DB) error {
	return tx.Save(a).Error
}

type AlgorithmResult struct {
	Uuid       string    `json:"uuid" gorm:"column:uuid;primary_key"`
	Name       string    `json:"name" gorm:"column:name"`
	Desc       string    `json:"desc" gorm:"column:desc"`
	Path       string    `json:"path" gorm:"column:path"`
	EntryPoint string    `json:"entry_point" gorm:"column:entry_point"`
	ExecBinary string    `json:"exec_binary" gorm:"column:exec_binary"`
	AuthorId   int       `json:"author_id" gorm:"column:author_id"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	Files      string    `json:"files" gorm:"column:files"`
	Username   string    `json:"username" gorm:"column:username"`
}
