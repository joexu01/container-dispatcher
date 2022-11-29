package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"gorm.io/gorm"
	"time"
)

type Task struct {
	Uuid          string    `json:"uuid" gorm:"column:uuid;primary_key"`
	TaskName      string    `json:"task_name" gorm:"column:task_name"`
	TaskDesc      string    `json:"task_desc" gorm:"column:task_desc"`
	AlgorithmUuid string    `json:"algorithm_uuid" gorm:"column:algorithm_uuid"`
	UserId        int       `json:"user_id" gorm:"column:user_id"`
	UploadedFiles string    `json:"uploaded_files" gorm:"column:uploaded_files"`
	Status        string    `json:"status" gorm:"column:status"`
	ImageName     string    `json:"image_name" gorm:"column:image_name"`
	ContainerId   string    `json:"container_id" gorm:"column:container_id"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at"`
}

func (t *Task) TableName() string {
	return "task"
}

func (t *Task) FindByUserID(_ *gin.Context, tx *gorm.DB, taskUuid string, userID int) (*TaskQueryResult, error) {
	var err error
	out := &TaskQueryResult{}

	if userID != 0 {

		err = tx.Select("task.uuid, task.task_name, task.task_desc, task.algorithm_uuid, task.user_id, task.uploaded_files, task.status, task.created_at, task.image_name, task.container_id, user.username, algorithm.name").
			Table(t.TableName()).
			Joins("join user on user.id = task.user_id").
			Joins("join algorithm on algorithm.uuid = task.algorithm_uuid").
			Where("user.id=? and task.uuid=?", userID, taskUuid).First(out).Error
	} else {
		err = tx.Model(out).Where(&Task{Uuid: taskUuid}).Scan(out).Error
	}

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (t *Task) Find(_ *gin.Context, tx *gorm.DB, taskUuid string) (*Task, error) {
	var err error
	out := &Task{}

	err = tx.Model(out).Where(&Task{Uuid: taskUuid}).Find(out).Error

	if err != nil {
		return nil, err
	}
	return out, nil
}

func (t *Task) PageList(_ *gin.Context, tx *gorm.DB, param *dto.TaskListQueryInput, userID int) (total int64, algoList []*TaskQueryResult, err error) {
	query := tx.Select("task.uuid, task.task_name, task.task_desc, task.algorithm_uuid, task.user_id, task.uploaded_files, task.status, task.created_at, task.image_name, task.container_id, user.username, algorithm.name").
		Table(t.TableName()).
		Joins("join user on user.id = task.user_id").
		Joins("join algorithm on algorithm.uuid = task.algorithm_uuid").
		Where("user.id=?", userID)

	offset := (param.PageNo - 1) * param.PageSize
	//Limit(param.PageSize).Offset(offset)
	if err = query.Limit(param.PageSize).Offset(offset).Order("task.created_at desc").Scan(&algoList).Error; err != gorm.ErrRecordNotFound && err != nil {
		return 0, nil, err
	}
	query.Limit(param.PageSize).Offset(offset).Count(&total)
	return total, algoList, nil
}

func (t *Task) Create(_ *gin.Context, tx *gorm.DB) error {
	return tx.Create(t).Error
}

func (t *Task) Update(_ *gin.Context, tx *gorm.DB) error {
	return tx.Save(t).Error
}

type TaskQueryResult struct {
	Uuid            string    `json:"uuid" gorm:"column:uuid;primary_key"`
	TaskName        string    `json:"task_name" gorm:"column:task_name"`
	TaskDesc        string    `json:"task_desc" gorm:"column:task_desc"`
	AlgorithmUuid   string    `json:"algorithm_uuid" gorm:"column:algorithm_uuid"`
	UserId          int       `json:"user_id" gorm:"column:user_id"`
	UploadedFiles   string    `json:"uploaded_files" gorm:"column:uploaded_files"`
	Status          string    `json:"status" gorm:"column:status"`
	ImageName       string    `json:"image_name" gorm:"column:image_name"`
	ContainerId     string    `json:"container_id" gorm:"column:container_id"`
	CreatedAt       time.Time `json:"created_at" gorm:"column:created_at"`
	ContainerStatus string    `json:"container_status"`
	//ContainerState *types.ContainerState `json:"container_state"`

	Username      string `json:"username" gorm:"column:username"`
	AlgorithmName string `json:"algorithm_name" gorm:"column:name"`
}

type TaskListWrapper struct {
	Total int64               `json:"total"`
	List  *[]*TaskQueryResult `json:"list"`
}
