package controller

import (
	bytes2 "bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	uuidPkg "github.com/google/uuid"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const GigaByte5 = 5 * 1024 * 1024 * 1024

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
	group.GET("/detail/:task_id", tc.TaskDetail)
	group.GET("/run/:task_id", tc.RunTaskTest)
	group.GET("/stop/:task_id", tc.StopTask)
	group.GET("/log/:task_id", tc.ShowTaskLog)
	group.GET("/remove/:task_id", tc.RemoveTask)

	group.StaticFS("/file/", gin.Dir(lib.GetStringConf("base.task_file.directory"), true))
	group.GET("/dir/:task_id", tc.GetTaskFiles)
	group.GET("/pack", tc.PackAndDownloadFiles)
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
	total, taskList, err := handler.PageList(c, db, params, user.Id)
	if err != nil {
		middleware.ResponseError(c, 2007, err)
		return
	}

	cli, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}
	defer func(Client *client.Client) {
		_ = Client.Close()
	}(cli.Client)

	for _, taskPointer := range taskList {
		t := taskPointer
		if t.Status == public.TaskStatusRemoved {
			continue
		}
		inspect, err := cli.Client.ContainerInspect(c, t.ContainerId)
		if err != nil {
			t.ContainerStatus = "removed or not created"
			continue
		}
		t.ContainerStatus = inspect.State.Status
	}

	out := &dao.TaskListWrapper{Total: total, List: &taskList}
	middleware.ResponseSuccess(c, out)
}

// TaskDetail godoc
// @Summary      当前用户任务列表
// @Description  当前用户任务列表
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/detail/:task_id [get]
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

	if (task.Status == public.TaskStatusContainerCreated || task.Status == public.TaskStatusTerminated) && task.ContainerId != "" {
		cli, err := lib.NewDockerClient("default")
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
			return
		}
		defer func(Client *client.Client) {
			_ = Client.Close()
		}(cli.Client)

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
// @Param        gpu      query      string   true   "GPU UUID"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/run/:task_id [get]
func (t *TaskController) RunTaskTest(c *gin.Context) {
	tId := c.Param("task_id")
	if tId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid task ID"), "")
		return
	}

	var resources *container.Resources

	request := container.DeviceRequest{
		Driver:       "nvidia",
		Count:        0,
		DeviceIDs:    nil,
		Capabilities: [][]string{{"gpu"}},
		Options:      nil,
	}

	var deviceIDs []string
	gpuUuids := c.Query("gpu")

	if gpuUuids == "none" {
		resources = &container.Resources{}
	} else {
		if gpuUuids == "" || len(gpuUuids) < 10 {
			request.Count = -1
			resources = &container.Resources{DeviceRequests: []container.DeviceRequest{request}}
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
			resources = &container.Resources{DeviceRequests: []container.DeviceRequest{request}}
		}
	}

	//log.Println("Device Request ----", request)

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
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

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
			Resources:     *resources,
		},
		nil, nil, task.Uuid)

	if err != nil {
		log.Println("Cmd ----", algorithm.ExecBinary, algorithm.EntryPoint)
		log.Println("Image Name ----", task.ImageName)
		log.Println("Binds ----", dirName+":/workspace")
		log.Println("Binds ----", dirName+":/workspace")

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

	middleware.ResponseSuccess(c, fmt.Sprintf("Container Created! %v", errMessage))
}

// StopTask godoc
// @Summary      终止任务容器运行
// @Description  终止任务容器运行
// @Tags         task
// @Produce      json
// @Param        gpu      query      string   true   "GPU UUID"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/stop/:task_id [get]
func (t *TaskController) StopTask(c *gin.Context) {
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

	// get docker client
	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

	timeout := time.Duration(time.Second)
	err = dockerClient.Client.ContainerStop(c, task.ContainerId, &timeout)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, errors.New("停止容器失败"), "")
		return
	}

	task.Status = public.TaskStatusTerminated
	err = task.Update(c, db)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}

	middleware.ResponseSuccess(c, "任务已终止，容器停止运行")
}

// RemoveTask godoc
// @Summary      删除任务
// @Description  删除任务
// @Tags         task
// @Produce      json
// @Param        gpu      query      string   true   "GPU UUID"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/remove/:task_id [get]
func (t *TaskController) RemoveTask(c *gin.Context) {
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

	var dockerOptsErr []error

	// get docker client
	if task.ContainerId != "" {
		dockerClient, err := lib.NewDockerClient("default")
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
			return
		}
		defer func(Client *client.Client) {
			err := Client.Close()
			if err != nil {
				log.Println("Handler Error:", err.Error())
			}
		}(dockerClient.Client)

		timeout := time.Duration(time.Second)
		err = dockerClient.Client.ContainerStop(c, task.ContainerId, &timeout)
		if err != nil {
			dockerOptsErr = append(dockerOptsErr, err)
		}

		err = dockerClient.Client.ContainerRemove(c, task.ContainerId, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   false,
			Force:         false,
		})
		if err != nil {
			dockerOptsErr = append(dockerOptsErr, err)
		}
	}

	baseDir := lib.GetStringConf("base.task_file.directory")
	dirName := baseDir + task.Uuid

	exists, err := public.PathExists(dirName)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	if exists {
		err = os.RemoveAll(dirName)
	}

	task.Status = public.TaskStatusRemoved
	err = task.Update(c, db)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}

	if len(dockerOptsErr) != 0 {
		var errStr string
		for _, e := range dockerOptsErr {
			errStr += e.Error()
		}

		middleware.ResponseSuccess(c, "任务已删除，但在删除过程中出现了如下错误："+errStr)
		return
	}
	middleware.ResponseSuccess(c, "任务已删除")
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
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

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
	defer func(containerLogs io.ReadCloser) {
		_ = containerLogs.Close()
	}(containerLogs)

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2009, err, "")
		return
	}

	bufLog := new(bytes2.Buffer)

	_, _ = stdcopy.StdCopy(bufLog, bufLog, containerLogs)

	//bytes2.NewReader()

	bytes, err := ioutil.ReadAll(bufLog)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2011, err, "")
		return
	}
	logsStr := string(bytes)

	split := strings.Split(logsStr, "\n")

	log.Println(split)

	middleware.ResponseSuccess(c, split)
	//c.Stream()
}

// GetTaskFiles godoc
// @Summary      返回任务文件夹文件列表
// @Description  返回任务文件夹文件列表
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/dir/:task_id [get]
func (t *TaskController) GetTaskFiles(c *gin.Context) {
	tId := c.Param("task_id")
	if tId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid task ID"), "")
		return
	}

	dirName := lib.GetStringConf("base.task_file.directory") + tId

	exists, err := public.PathExists(dirName)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2008, err, "")
		return
	}

	if !exists {
		middleware.ResponseSuccess(c, &dto.DirFileInfo{})
	}

	tree := &public.FileTree{
		Label:    "root",
		Filepath: "",
		Children: nil,
	}

	public.GetFilesWithDirInfo(dirName, "", tree)

	middleware.ResponseSuccess(c, tree)
}

// PackAndDownloadFiles godoc
// @Summary      返回任务文件夹文件列表
// @Description  返回任务文件夹文件列表
// @Tags         task
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /task/pack?filepath [get]
func (t *TaskController) PackAndDownloadFiles(c *gin.Context) {
	dirPath := c.Query("filepath")
	if len(dirPath) == 0 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("无效的文件夹"), "")
		return
	}

	dirAbsPath, err := filepath.Abs(dirPath)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("无效的文件夹"), "")
		return
	}

	if !strings.HasPrefix(dirAbsPath, lib.GetStringConf("base.task_file.directory")) {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("无效的文件夹"), "")
		return
	}

	stat, err := os.Stat(dirAbsPath)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}
	if !stat.IsDir() {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2002, errors.New("文件夹不存在"), "")
		return
	}

	dirSizeSum, err := public.DirSizeSum(dirAbsPath)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}

	if dirSizeSum > GigaByte5 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2004, errors.New("文件夹文件未压缩大小超过5GiB"), "")
		return
	}

	now := time.Now().UnixNano()
	dirName := strings.Split(dirAbsPath, `/`)
	if len(dirName) < 1 {
		dirName = []string{strconv.FormatInt(now, 10)}
	}
	zipFileName := lib.GetStringConf("base.task_file.directory") + dirName[len(dirName)-1] + strconv.FormatInt(now, 10) + ".zip"

	err = public.Zip(zipFileName, dirAbsPath)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}
	defer os.Remove(zipFileName)

	c.Header("Content-Disposition", "attachment; filename="+dirName[len(dirName)-1]+strconv.FormatInt(now, 10)+".zip")
	c.File(zipFileName)
}
