package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	uuidPkg "github.com/google/uuid"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"net/http"
	"os"
	"strings"
	"time"
)

type AlgorithmController struct{}

func AlgorithmControllerRegister(group *gin.RouterGroup) {
	ac := &AlgorithmController{}

	group.POST("/upload", ac.AlgorithmUploadNew)
	group.POST("/new", ac.NewAttackAlgorithm)
	group.GET("/list", ac.ListAttackAlgorithms)
}

// AlgorithmUploadNew godoc
// @Summary      上传攻击所需要的文件
// @Description  上传攻击所需要的文件
// @Tags         algorithm
// @Produce      json
// @Param     at_id      query      string   true   "攻击算法的ID"
// @Param     upload[]   formData   file     true   "上传的文件"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /algorithm/upload [post]
func (a *AlgorithmController) AlgorithmUploadNew(c *gin.Context) {
	attackID := c.Query("at_id")
	if attackID == "" || len(attackID) == 0 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("bad request: param at_id(attack algorithm id) is missing"), "")
		return
	}

	//baseDir := lib.GetStringConf("base.attack_file.directory")

	//uuid := uuidPkg.NewString()
	attackID = strings.TrimSpace(attackID)
	//dirName := baseDir + attackID

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	algo := &dao.Algorithm{Uuid: attackID}
	algorithm, err := algo.Find(c, db, algo)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}

	dirName := algorithm.Path

	exists, err := public.PathExists(dirName)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	if !exists {
		err := os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			middleware.ResponseWithCode(c, http.StatusInternalServerError, 2002, err, "")
			return
		}
	}

	form, _ := c.MultipartForm()
	files := form.File["file"]

	if len(files) == 0 {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2003, errors.New("no files uploaded, check your form"), "")
		return
	}

	filenames := algorithm.Files

	for _, file := range files {
		fmt.Println(file.Filename)
		filenames = filenames + file.Filename + ";"
		_ = c.SaveUploadedFile(file, dirName+`/`+file.Filename)

	}

	algorithm.Files = filenames
	err = algorithm.Update(c, db)

	middleware.ResponseSuccess(c, "uploaded files: "+filenames)
}

// NewAttackAlgorithm godoc
// @Summary      新建攻击算法
// @Description  新建攻击算法
// @Tags         algorithm
// @Produce      json
// @Param        struct body dto.AlgorithmParams true "上传新算法的必备参数"
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /algorithm/new [post]
func (a *AlgorithmController) NewAttackAlgorithm(c *gin.Context) {
	params := &dto.AlgorithmParams{}
	err := params.BindValidParam(c)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, err, "")
		return
	}

	baseDir := lib.GetStringConf("base.attack_file.directory")

	uuid := uuidPkg.NewString()
	dirName := baseDir + uuid

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

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	algo := &dao.Algorithm{
		Uuid:             uuid,
		Name:             params.Name,
		Desc:             params.Desc,
		Path:             dirName,
		EntryPoint:       params.EntryPoint,
		ExecBinary:       params.ExecBinary,
		AuthorId:         user.Id,
		CreatedAt:        time.Now(),
		Files:            "",
		DefaultImageName: params.DefaultImageName,
	}

	err = algo.Create(c, db)

	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}

	err = os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	type id struct {
		AtID string `json:"at_id"`
	}

	middleware.ResponseSuccess(c, &id{uuid})
}

// ListAttackAlgorithms godoc
// @Summary      攻击算法列表
// @Description  攻击算法列表
// @Tags         algorithm
// @Produce      json
// @Success      200  {object}  middleware.Response{data=string} "success"
// @Failure      500  {object}  middleware.Response
// @Router       /algorithm/list [get]
func (a *AlgorithmController) ListAttackAlgorithms(c *gin.Context) {
	params := &dto.AlgorithmListQueryInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2001, err, "")
		return
	}

	handler := &dao.Algorithm{}
	total, algoList, err := handler.PageList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	out := &dao.AlgorithmWrapper{Total: total, List: &algoList}

	middleware.ResponseSuccess(c, out)
}
