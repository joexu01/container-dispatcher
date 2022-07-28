package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"strconv"
	"strings"
	"time"
)

type ImageController struct{}

func ImageControllerRegister(group *gin.RouterGroup) {
	img := &ImageController{}
	group.GET("/list", img.ImagesList)
}

// ImagesList godoc
// @Summary      获取镜像列表
// @Description  获取镜像列表
// @Tags         image
// @Produce      json
// @Success      200  {object}  middleware.Response{data=dto.ImageList} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /image/list [get]
func (i *ImageController) ImagesList(c *gin.Context) {
	client, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	images, err := client.ListAllImages()
	if err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}

	var imageList []*dto.ImageInfo
	for _, image := range images {
		i := image
		repo := ""
		for _, r := range i.RepoTags {
			repo = repo + r
		}
		id := i.ID
		idSplit := strings.Split(id, ":")
		if len(idSplit) >= 2 {
			id = idSplit[1]
			rs := []rune(id)
			id = string(rs[0:12])
		}
		createdSec := strconv.Itoa(int(i.Created)) + "s"
		duration, err := time.ParseDuration(createdSec)
		if err == nil {
			createdSec = duration.String()
		}
		s := i.Size / (1024 * 1024)
		size := strconv.Itoa(int(s)) + "MB"

		imgInfo := &dto.ImageInfo{
			Repo:    repo,
			ImageId: id,
			Created: createdSec,
			Size:    size,
		}
		imageList = append(imageList, imgInfo)
	}

	out := &dto.ImageList{List: imageList}

	middleware.ResponseSuccess(c, out)
}
