package controller

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"net/http"
	"time"
)

type UserLogoutController struct{}

func UserLogoutRegister(group *gin.RouterGroup) {
	user := &UserLogoutController{}
	group.GET("/logout", user.UserLogout)
	group.POST("/register", user.UserRegister)
	group.GET("/list", user.UserList)
}

// UserLogout godoc
// @Summary      用户登出
// @Description  就是用户登出呗
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  middleware.Response
// @Failure      500  {object}  middleware.Response
// @Router       /user/logout [get]
func (u *UserLogoutController) UserLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(public.UserSessionKey)
	_ = session.Save()
	middleware.ResponseSuccess(c, "")
}

// UserRegister godoc
// @Summary      用户注册（仅限管理员）
// @Description  用户注册
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        struct body dto.NewUserInput true "新建用户输入"
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /user/register [post]
func (u *UserLogoutController) UserRegister(c *gin.Context) {
	session := sessions.Default(c)
	sessStr, ok := session.Get(public.UserSessionKey).(string)
	if !ok {
		middleware.ResponseError(c, 2001, errors.New("login before execute this operation"))
		return
	}

	sessInfo := &dto.UserSessionInfo{}

	err := json.Unmarshal([]byte(sessStr), sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, errors.New("cannot find login record"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}

	search := &dao.User{Id: sessInfo.Id}
	user, err := search.Find(c, db, search)
	if err != nil || user.UserRole != public.UserRoleAdmin {
		middleware.ResponseError(c, 2004, errors.New("an error occured during fetching user profile"))
		return
	}

	newUserInput := &dto.NewUserInput{}
	err = newUserInput.BindValidParam(c)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "")
		return
	}

	pwd, err := public.GeneratePwdHash([]byte(newUserInput.RawPassword))
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, err, "")
		return
	}

	newUser := &dao.User{
		Id:        0,
		Username:  newUserInput.Username,
		Password:  pwd,
		Email:     newUserInput.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsDelete:  0,
		UserRole:  newUserInput.UserRole,
	}

	err = newUser.Save(c, db)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, err, "")
		return
	}

	middleware.ResponseSuccess(c, "ok")
}

// UserList godoc
// @Summary      用户列表
// @Description  用户列表
// @Tags         user
// @Produce      json
// @Param     page_size   query   int   false   "page size"
// @Param     page_no     query   int   false   "page no"
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /user/list [get]
func (u *UserLogoutController) UserList(c *gin.Context) {

	params := &dto.UserListQueryInput{}
	if err := params.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}

	session := sessions.Default(c)
	sessStr, ok := session.Get(public.UserSessionKey).(string)
	if !ok {
		middleware.ResponseError(c, 2001, errors.New("login before execute this operation"))
		return
	}

	sessInfo := &dto.UserSessionInfo{}

	err := json.Unmarshal([]byte(sessStr), sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, errors.New("cannot find login record"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "")
		return
	}

	search := &dao.User{Id: sessInfo.Id}
	user, err := search.Find(c, db, search)
	if err != nil || user.UserRole != public.UserRoleAdmin {
		middleware.ResponseError(c, 2004, errors.New("an error occurred during fetching user profile"))
		return
	}

	handler := &dao.User{}
	_, userList, err := handler.PageList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	middleware.ResponseSuccess(c, userList)
}
