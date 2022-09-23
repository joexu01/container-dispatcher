package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	uuidPkg "github.com/google/uuid"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	uuidReg  = regexp.MustCompile(`^GPU-*`)
	indexReg = regexp.MustCompile(`^[0-9]+`)
)

type TaskController struct{}

func TaskControllerRegister(group *gin.RouterGroup) {
	tc := &TaskController{}
	group.POST("/upload", tc.TaskUploadFiles)
	group.POST("/new", tc.NewTask)
	group.GET("/list", tc.ListTasks)
	group.GET("/:task_id", tc.TaskDetail)
	group.GET("/run/:task_id", tc.RunTaskTest)
	group.GET("/log/:task_id", tc.ShowTaskLog)
}

// TaskUploadFiles godoc
// @Summary      上传用户模型文件
// @Description  上传用户模型文件
// @Tags         task
// @Produce      json
// @Param     task_id      query    string    true   "Task的ID"
// @Param     upload[]   formData   file     true   "上传的文件"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/upload [post]
func (t *TaskController) TaskUploadFiles(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" || len(taskID) == 0 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("bad request: param task_id(attack task id) is missing"), "")
		return
	}
	fmt.Println(taskID)

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	taskSearch := &dao.Task{Uuid: taskID}
	task, err := taskSearch.Find(c, db, taskID)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2006, errors.New("bad request: task not found"), "")
		return
	}

	if task == nil || task.Uuid == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2007, errors.New("bad request: task not found"), "")
		return
	}

	baseDir := lib.GetStringConf("base.task_file.directory")
	dirName := baseDir + task.Uuid

	exists, err := public.PathExists(dirName)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2008, err, "")
		return
	}

	if !exists {
		err := os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2010, err, "")
			return
		}
	}

	form, _ := c.MultipartForm()
	files := form.File["file"]

	if len(files) == 0 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2012, errors.New("no files uploaded, check your form"), "")
		return
	}

	var filenames string

	for _, file := range files {
		fmt.Println(file.Filename)
		filenames = filenames + file.Filename + ";"
		_ = c.SaveUploadedFile(file, dirName+`/`+file.Filename)

	}

	task.UploadedFiles = filenames
	err = task.Update(c, db)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2014, err, "")
		return
	}

	middleware.ResponseSuccess(c, "uploaded files: "+filenames)
}

// NewTask godoc
// @Summary      新建任务
// @Description  新建任务
// @Tags         task
// @Produce      json
// @Param        struct body dto.NewTaskParameters true "新建任务参数"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/new [post]
func (t *TaskController) NewTask(c *gin.Context) {
	params := &dto.NewTaskParameters{}
	err := params.BindValidParam(c)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, err, "")
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, err, "")
		return
	}

	algoSearch := &dao.Algorithm{Uuid: params.AlgorithmUuid}
	algo, err := algoSearch.Find(c, db, algoSearch)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2004, err, "")
		return
	}

	if algo == nil || algo.Uuid == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2005, errors.New("algorithm not found"), "")
		return
	}

	session := sessions.Default(c)
	userSessInfo := session.Get(public.UserSessionKey).(string)
	if len(userSessInfo) == 0 {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, errors.New("invalid user session info"), "")
		return
	}

	user := &dto.UserSessionInfo{}
	err = json.Unmarshal([]byte(userSessInfo), user)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2008, err, "")
		return
	}

	uuid := uuidPkg.NewString()

	task := &dao.Task{
		Uuid:          uuid,
		TaskName:      params.Name,
		TaskDesc:      params.Desc,
		AlgorithmUuid: params.AlgorithmUuid,
		UserId:        user.Id,
		UploadedFiles: "",
		Status:        public.TaskStatusReady,
		ImageName:     algo.DefaultImageName,
		ContainerId:   "",
		CreatedAt:     time.Now(),
	}

	baseDir := lib.GetStringConf("base.task_file.directory")
	dirName := baseDir + uuid

	err = task.Create(c, db)

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2010, err, "")
		return
	}

	err = os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2012, err, "")
		return
	}

	type id struct {
		TaskId string `json:"task_id"`
	}

	middleware.ResponseSuccess(c, &id{uuid})

}

// ListTasks godoc
// @Summary      当前用户任务列表
// @Description  当前用户任务列表
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/list [get]
func (t *TaskController) ListTasks(c *gin.Context) {
	params := &dto.TaskListQueryInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, err, "")
		return
	}

	session := sessions.Default(c)
	userSessInfo := session.Get(public.UserSessionKey).(string)
	if len(userSessInfo) == 0 {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, errors.New("invalid user session info"), "")
		return
	}

	user := &dto.UserSessionInfo{}
	err = json.Unmarshal([]byte(userSessInfo), user)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, err, "")
		return
	}

	handler := &dao.Task{}
	_, taskList, err := handler.PageList(c, db, params, user.Id)
	if err != nil {
		middleware.ResponseError(c, 2007, err)
		return
	}

	middleware.ResponseSuccess(c, taskList)
}

// TaskDetail godoc
// @Summary      当前用户任务列表
// @Description  当前用户任务列表
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/:task_id [get]
func (t *TaskController) TaskDetail(c *gin.Context) {
	tId := c.Param("task_id")
	if tId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid task ID"), "")
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	session := sessions.Default(c)
	userSessInfo := session.Get(public.UserSessionKey).(string)
	if len(userSessInfo) == 0 {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, errors.New("invalid user session info"), "")
		return
	}

	user := &dto.UserSessionInfo{}
	err = json.Unmarshal([]byte(userSessInfo), user)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}

	search := &dao.Task{Uuid: tId}
	task, err := search.FindByUserID(c, db, tId, user.Id)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}
	taskInDb, err := search.Find(c, db, tId)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2009, err, "")
		return
	}

	if task.Status == public.TaskStatusContainerCreated && task.ContainerId != "" {
		cli, err := lib.NewDockerClient("default")
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
			return
		}

		inspect, err := cli.Client.ContainerInspect(c, task.ContainerId)
		if err != nil {
			middleware.ResponseSuccess(c, task)
			return
		}

		task.ContainerStatus = inspect.State.Status
		taskInDb.Status = inspect.State.Status
	}
	middleware.ResponseSuccess(c, task)
}

// RunTaskTest godoc
// @Summary      测试-运行任务容器
// @Description  测试-运行任务容器
// @Tags         task
// @Produce      json
// @Param        gpu      query      string   true   "攻击算法的ID"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/run/:task_id [get]
func (t *TaskController) RunTaskTest(c *gin.Context) {
	tId := c.Param("task_id")
	if tId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid task ID"), "")
		return
	}

	request := container.DeviceRequest{
		Driver:       "nvidia",
		Count:        0,
		DeviceIDs:    nil,
		Capabilities: [][]string{{"gpu"}},
		Options:      nil,
	}

	var deviceIDs []string
	gpuUuids := c.Query("gpu")
	if gpuUuids == "" || len(gpuUuids) < 10 {
		request.Count = -1
	} else {
		uuids := strings.Split(gpuUuids, "_")
		for _, uuid := range uuids {
			u := uuid
			if !uuidReg.MatchString(u) && !indexReg.MatchString(u) {
				continue
			}
			deviceIDs = append(deviceIDs, u)
		}
		request.DeviceIDs = deviceIDs
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	// get task info

	searchTask := &dao.Task{}
	task, err := searchTask.Find(c, db, tId)
	if err != nil || task.Uuid == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2002, errors.New("task not found"), "")
		return
	}

	baseDir := lib.GetStringConf("base.task_file.directory")
	dirName := baseDir + task.Uuid

	exists, _ := public.PathExists(dirName)
	if !exists {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, errors.New("task folder not found"), "")
		return
	}

	// get algorithm detail

	searchAlgorithm := &dao.Algorithm{Uuid: task.AlgorithmUuid}
	algorithm, err := searchAlgorithm.Find(c, db, searchAlgorithm)
	if err != nil || algorithm.Uuid == "" {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, errors.New("task folder not found"), "")
		return
	}

	// copy files from algo dir to task dir

	fileInfos, err := ioutil.ReadDir(algorithm.Path)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, errors.New("task folder not found"), "")
		return
	}

	for _, file := range fileInfos {
		_, err := public.CopyFile(algorithm.Path+`/`+file.Name(), dirName+`/`+file.Name())
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, err, "")
			return
		}
	}

	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2007, err, "")
		return
	}

	resp, err := dockerClient.Client.ContainerCreate(c,
		&container.Config{
			Cmd:         []string{algorithm.ExecBinary, algorithm.EntryPoint},
			Image:       task.ImageName,
			WorkingDir:  "/workspace",
			Entrypoint:  nil,
			StopSignal:  "",
			StopTimeout: nil,
			Shell:       nil,
		},
		&container.HostConfig{
			Binds:         []string{dirName + ":/workspace"},
			RestartPolicy: container.RestartPolicy{},
			ConsoleSize:   [2]uint{},
			Resources:     container.Resources{DeviceRequests: []container.DeviceRequest{request}},
		},
		nil, nil, task.Uuid)

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2008, err, "")
		return
	}

	if err = dockerClient.Client.ContainerStart(c, resp.ID, types.ContainerStartOptions{}); err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2009, err, "")
		return
	}

	task.Status = public.TaskStatusContainerCreated
	task.ContainerId = resp.ID[0:12]

	var errMessage string

	err = task.Update(c, db)
	if err != nil {
		errMessage += err.Error()
	}

	middleware.ResponseSuccess(c, fmt.Sprintf("Container Created, while handling task, some error may occur: %v", errMessage))
}

// ShowTaskLog godoc
// @Summary      查看容器运行输出
// @Description  查看容器运行输出 - docker logs
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/log/:task_id [get]
func (t *TaskController) ShowTaskLog(c *gin.Context) {
	tId := c.Param("task_id")
	if tId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid task ID"), "")
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	// get task info

	searchTask := &dao.Task{}
	task, err := searchTask.Find(c, db, tId)
	if err != nil || task.Uuid == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2002, errors.New("task not found"), "")
		return
	}

	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2007, err, "")
		return
	}

	containerLogs, err := dockerClient.Client.ContainerLogs(c, task.ContainerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Since:      "",
		Until:      "",
		Timestamps: false,
		Follow:     false,
		Tail:       "100",
		Details:    false,
	})

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2009, err, "")
		return
	}

	bytes, err := ioutil.ReadAll(containerLogs)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2011, err, "")
		return
	}
	_ = containerLogs.Close()
	logsStr := string(bytes)
	middleware.ResponseSuccess(c, logsStr)
}
