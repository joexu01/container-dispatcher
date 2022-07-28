package controller

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
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

	for _, container := range containers {
		c := container
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
		if container, ok := containerInfoMap[uContainer.ContainerId]; ok {

			names := ""
			if len(container.Names) != 1 {
				for _, n := range container.Names {
					na := n
					na = strings.Replace(na, "/", "", -1)
					names = names + na
					names = names + " | "
				}
			} else {
				na := strings.Replace(container.Names[0], "/", "", -1)
				names = na
			}

			portStr := ""
			for _, port := range container.Ports {
				p := port
				ps := fmt.Sprintf("IP Addr:%s;Host Port:%d;Container Port: %d/%s | ", p.IP, p.PublicPort, p.PrivatePort, p.Type)
				portStr = portStr + ps
			}
			createdSec := strconv.Itoa(int(container.Created)) + "s"
			duration, err := time.ParseDuration(createdSec)
			if err == nil {
				createdSec = duration.String()
			}

			ucInfo := &dto.UserContainerInfoFull{
				Id:          u.Id,
				Username:    u.Username,
				ContainerId: u.ContainerId,
				Image:       container.Image,
				Command:     container.Command,
				Created:     createdSec,
				Status:      container.Status,
				Ports:       portStr,
				Name:        names,
			}
			userConDtoList = append(userConDtoList, ucInfo)
		}
	}

	middleware.ResponseSuccess(c, &dto.UserContainerInfoList{List: userConDtoList})
}

func (t *ContainerController) RunContainer(c *gin.Context) {
	//client, err := lib.NewDockerClient("default")
	//if err != nil {
	//	middleware.ResponseError(c, 2000, err)
	//	return
	//}
}
