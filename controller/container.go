package controller

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"net/http"
	"strconv"
	"time"
)

type ContainerController struct{}

func ContainerControllerRegister(group *gin.RouterGroup) {
	ctn := &ContainerController{}
	group.GET("/list", ctn.ContainerList)
}

// ContainerList godoc
// @Summary      获取我的容器列表
// @Description  获取我的容器列表
// @Tags         container
// @Produce      json
// @Param     page_size   query   int   false   "page size"
// @Param     page_no     query   int   false   "page no"
// @Success      200  {object}  middleware.Response "success"
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

	containerInfoMap := make(map[string]*types.Container)

	for _, container := range containers {
		c := &container
		contIdRaw := []rune(c.ID)
		contId := string(contIdRaw[0:12])

		containerInfoMap[contId] = c
	}

	//var userConDtoList []*dto.UserContainerInfo

	for _, uContainer := range userContainerList {
		u := &uContainer

		if container, ok := containerInfoMap[uContainer.ContainerId]; ok {
			u.Command = container.Command

			names := ""
			for _, n := range container.Names {
				na := n
				names = names + na
				names = names + " | "
			}
			u.Name = names

			u.Image = container.Image

			portStr := ""
			for _, port := range container.Ports {
				p := port
				ps := fmt.Sprintf("|IP Addr: %s; Container Port: %d; Host Port: %d; Type: %s |", p.IP, p.PrivatePort, p.PublicPort, p.Type)
				portStr = portStr + ps
			}
			u.Ports = portStr

			u.Status = container.Status

			createdSec := strconv.Itoa(int(container.Created)) + "s"
			duration, err := time.ParseDuration(createdSec)
			if err == nil {
				createdSec = duration.String()
			}
			u.Created = createdSec
		}
	}

	middleware.ResponseSuccess(c, userContainerList)
}

func (t *ContainerController) RunContainer(c *gin.Context) {
	//client, err := lib.NewDockerClient("default")
	//if err != nil {
	//	middleware.ResponseError(c, 2000, err)
	//	return
	//}
}
