package controller

import (
	"bufio"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ImageController struct{}

func ImageControllerRegister(group *gin.RouterGroup) {
	img := &ImageController{}
	group.GET("/list", img.ImagesList)
	group.POST("/upload", img.UploadImage)
	group.GET("/remove/:image_id", img.RemoveImage)
	group.GET("/pull", img.PullImage)
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
	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

	images, err := dockerClient.ListAllImages()
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
		cTime := time.Unix(i.Created, 0).String()
		//createdSec := strconv.Itoa(int(i.Created)) + "s"
		//duration, err := time.ParseDuration(createdSec)
		//if err == nil {
		//	createdSec = duration.String()
		//}
		s := i.Size / (1024 * 1024)
		size := strconv.Itoa(int(s)) + "MB"

		imgInfo := &dto.ImageInfo{
			Repo:    repo,
			ImageId: id,
			Created: cTime,
			Size:    size,
		}
		imageList = append(imageList, imgInfo)
	}

	out := &dto.ImageList{List: imageList, Total: int64(len(imageList))}

	middleware.ResponseSuccess(c, out)
}

// UploadImage godoc
// @Summary      上传镜像
// @Description  上传镜像
// @Tags         image
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /image/upload [post]
func (i *ImageController) UploadImage(c *gin.Context) {
	imageDir := lib.GetStringConf("base.image_file.directory")

	exists, err := public.PathExists(imageDir)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	if !exists {
		err := os.Mkdir(imageDir, os.ModePerm)
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, err, "")
			return
		}
	}

	file, _ := c.FormFile("file")

	if !strings.HasSuffix(file.Filename, ".tar") {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, errors.New("无效的文件后缀，请上传以 .tar 结尾的镜像打包文件"), "")
		return
	}

	imgFilename := imageDir + file.Filename

	_ = c.SaveUploadedFile(file, imgFilename)

	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

	imageFile, err := os.Open(imgFilename)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	imageReader := bufio.NewReader(imageFile)

	loadResponse, err := dockerClient.Client.ImageLoad(c, imageReader, true)
	if err != nil {
		middleware.ResponseError(c, 2006, err)
		return
	}
	_ = imageFile.Close()
	_ = os.Remove(imgFilename)

	if loadResponse.JSON {
		middleware.ResponseSuccess(c, "镜像上传成功")
		return
	}

	middleware.ResponseError(c, 2007, errors.New("镜像保存失败，请检查镜像文件"))
}

// RemoveImage godoc
// @Summary      上传镜像
// @Description  上传镜像
// @Tags         image
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /image/remove/:image_id [get]
func (i *ImageController) RemoveImage(c *gin.Context) {
	imgID := c.Param("image_id")
	if imgID == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("无效的 Image ID"), "")
		return
	}
	imgID = strings.TrimSpace(imgID)

	dockerClient, err := lib.NewDockerClient("default")
	if err != nil {
		middleware.ResponseError(c, 2004, err)
		return
	}
	defer func(Client *client.Client) {
		err := Client.Close()
		if err != nil {
			log.Println("UploadImage Handler Error:", err.Error())
		}
	}(dockerClient.Client)

	deleteResponseItems, err := dockerClient.Client.ImageRemove(c, imgID, types.ImageRemoveOptions{
		Force:         false,
		PruneChildren: false,
	})
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}
	log.Printf("%+v\n", deleteResponseItems)

	middleware.ResponseSuccess(c, "已经删除镜像")
}

// PullImage godoc
// @Summary      上传镜像
// @Description  上传镜像
// @Tags         image
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /image/pull?image=xxx&tag=x.x.x [get]
func (i *ImageController) PullImage(c *gin.Context) {
	imageName := c.Query("image")
	tag := c.Query("tag")

	if len(imageName) < 1 || len(tag) < 1 {
		middleware.ResponseError(c, 2001, errors.New("请求参数错误，请检查镜像名称和tag"))
		return
	}

	imageStr := imageName + ":" + tag

	go func(imageRef string) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 60*time.Minute)
		defer cancelFunc()

		go func() {
			dockerClient, err := lib.NewDockerClient("default")
			if err != nil {
				return
			}
			defer func(Client *client.Client) {
				_ = Client.Close()
			}(dockerClient.Client)

			_, _ = dockerClient.Client.ImagePull(ctx, imageRef, types.ImagePullOptions{
				All:           false,
				RegistryAuth:  "",
				PrivilegeFunc: nil,
				Platform:      "",
			})
		}()

		select {
		case <-ctx.Done():
			return
		}

	}(imageStr)

	middleware.ResponseSuccess(c, "正在拉取镜像，网络原因可能拉取失败，本次拉取超时时间1小时")
}
