package controller

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/log"
	"github.com/joexu01/container-dispatcher/middleware"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ContainerController struct{}

func ContainerControllerRegister(group *gin.RouterGroup) {
	ctn := &ContainerController{}
	group.GET("/list", ctn.ContainerList)
	group.POST("/run", ctn.RunContainerTest)
}

// ContainerList godoc
// @Summary      获取我的容器列表
// @Description  获取我的容器列表
// @Tags         container
// @Produce      json
// @Param     page_size   query   int   false   "page size"
// @Param     page_no     query   int   false   "page no"
// @Success      200  {object}  middleware.Response{data=dto.UserContainerInfoList} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /container/list [get]
func (t *ContainerController) ContainerList(c *gin.Context) {
	params := &dto.UserContainerListQueryInput{}
	err := params.BindValidParam(c)
	if err != nil {
		middleware.ResponseError(c, 2000, errors.New("parameters are missing"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	cu := &dao.ContainerUser{}
	userContainerList, err := cu.GetContainerList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}

	cli, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}
	containers, err := cli.ListContainers(true)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	containerInfoMap := make(map[string]types.Container)

	for _, cont := range containers {
		c := cont
		contIdRaw := []rune(c.ID)
		contId := string(contIdRaw[0:12])

		containerInfoMap[contId] = c
	}

	for k, v := range containerInfoMap {
		log.Debug(" Map: key: %+v ; val: %+v\n", k, v)
	}

	var userConDtoList []*dto.UserContainerInfoFull

	for _, uContainer := range userContainerList {
		u := uContainer
		if con, ok := containerInfoMap[uContainer.ContainerId]; ok {

			names := ""
			if len(con.Names) != 1 {
				for _, n := range con.Names {
					na := n
					na = strings.Replace(na, "/", "", -1)
					names = names + na
					names = names + " | "
				}
			} else {
				na := strings.Replace(con.Names[0], "/", "", -1)
				names = na
			}

			portStr := ""
			for _, port := range con.Ports {
				p := port
				ps := fmt.Sprintf("IP Addr:%s;Host Port:%d;Container Port: %d/%s | ", p.IP, p.PublicPort, p.PrivatePort, p.Type)
				portStr = portStr + ps
			}
			createdSec := strconv.Itoa(int(con.Created)) + "s"
			duration, err := time.ParseDuration(createdSec)
			if err == nil {
				createdSec = duration.String()
			}

			ucInfo := &dto.UserContainerInfoFull{
				Id:          u.Id,
				Username:    u.Username,
				ContainerId: u.ContainerId,
				Image:       con.Image,
				Command:     con.Command,
				Created:     createdSec,
				Status:      con.Status,
				Ports:       portStr,
				Name:        names,
			}
			userConDtoList = append(userConDtoList, ucInfo)
		}
	}

	middleware.ResponseSuccess(c, &dto.UserContainerInfoList{List: userConDtoList})
}

// RunContainerTest godoc
// @Summary      测试-运行任务容器
// @Description  测试-运行任务容器
// @Tags         container
// @Produce      json
// @Param        struct body dto.RunContainerParams true "运行容器的必备参数"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /container/run [post]
func (t *ContainerController) RunContainerTest(c *gin.Context) {
	params := &dto.RunContainerParams{}
	err := params.BindValidParam(c)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, err, "")
		return
	}

	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	resp, err := dockerClient.Client.ContainerCreate(c,
		&container.Config{
			Cmd:         []string{"/bin/bash", "/workspace/start.sh"},
			Image:       params.ImageName,
			WorkingDir:  "/workspace",
			Entrypoint:  nil,
			StopSignal:  "",
			StopTimeout: nil,
			Shell:       nil,
		},
		&container.HostConfig{
			Binds:         params.DirBinds,
			RestartPolicy: container.RestartPolicy{},
			ConsoleSize:   [2]uint{},
			Resources: container.Resources{DeviceRequests: []container.DeviceRequest{{
				Driver:       "nvidia",
				Count:        0,
				DeviceIDs:    params.GpuUuids,
				Capabilities: [][]string{{"gpu"}},
				Options:      nil,
			}}},
		},
		nil, nil, params.ContainerName)

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, err, "")
		return
	}

	if err = dockerClient.Client.ContainerStart(c, resp.ID, types.ContainerStartOptions{}); err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}

	middleware.ResponseSuccess(c, "Container Created.")
}
