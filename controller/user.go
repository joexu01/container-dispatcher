package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/joexu01/container-dispatcher/dao"
	"github.com/joexu01/container-dispatcher/dto"
	"github.com/joexu01/container-dispatcher/lib"
	"github.com/joexu01/container-dispatcher/middleware"
	"github.com/joexu01/container-dispatcher/public"
	"log"
	"net/http"
	"strconv"
	"time"
)

type UserLogoutController struct{}

func UserLogoutRegister(group *gin.RouterGroup) {
	user := &UserLogoutController{}
	group.GET("/logout", user.UserLogout)
	group.POST("/register", user.UserRegister)
	group.GET("/list", user.UserList)
	group.GET("/delete/:user_id", user.UserDelete)
	group.GET("/me", user.UserMe)

	group.GET("/debug/get", user.PrintGetRequest)
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
	middleware.ResponseSuccess(c, "已登出")
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
		middleware.ResponseError(c, 2004, errors.New("an error occurred during fetching user profile"))
		return
	}

	newUserInput := &dto.NewUserInput{}
	err = newUserInput.BindValidParam(c)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2004, err, "请求错误，请检查POST数据的Body")
		return
	}

	// 查找是否有用户名重名

	isExist := &dao.User{Username: newUserInput.Username}
	isExist, err = isExist.Find(c, db, isExist)
	log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!", isExist, err)
	if err != nil || isExist.Id != 0 {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2005, errors.New("用户名已存在 或者 数据库访问失败"), "用户名已存在 或者 数据库访问失败")
		return
	}

	pwd, err := public.GeneratePwdHash([]byte(newUserInput.RawPassword))
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, err, "内部错误")
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
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2007, err, "")
		return
	}

	middleware.ResponseSuccess(c, "用户创建成功")
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
	total, userList, err := handler.PageList(c, db, params)
	if err != nil {
		middleware.ResponseError(c, 2005, err)
		return
	}

	out := &dao.UserListWrapper{
		Total: total,
		List:  &userList,
	}

	middleware.ResponseSuccess(c, out)
}

// UserDelete godoc
// @Summary      用户删除（仅限管理员）
// @Description  用户删除
// @Tags         user
// @Produce      json
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /user/delete/:user_id [get]
func (u *UserLogoutController) UserDelete(c *gin.Context) {
	uId := c.Param("user_id")
	if uId == "" {
		middleware.ResponseWithCode(c, http.StatusBadRequest, 2000, errors.New("invalid user ID"), "")
		return
	}

	userId, err := strconv.Atoi(uId)
	if err != nil {
		middleware.ResponseError(c, 2000, errors.New("invalid user ID"))
		return
	}

	session := sessions.Default(c)
	sessStr, ok := session.Get(public.UserSessionKey).(string)
	if !ok {
		middleware.ResponseError(c, 2001, errors.New("login before execute this operation"))
		return
	}

	sessInfo := &dto.UserSessionInfo{}

	err = json.Unmarshal([]byte(sessStr), sessInfo)
	if err != nil {
		middleware.ResponseError(c, 2002, errors.New("cannot find login record"))
		return
	}

	db, err := lib.GetGormPool("default")
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2003, err, "cannot get db pool")
		return
	}

	search := &dao.User{Id: sessInfo.Id}
	user, err := search.Find(c, db, search)
	if err != nil || user.UserRole != public.UserRoleAdmin {
		middleware.ResponseError(c, 2004, errors.New("an error occurred during fetching user profile"))
		return
	}

	userSearch := &dao.User{Id: userId}
	userTobeDeleted, err := search.Find(c, db, userSearch)
	if err != nil {
		middleware.ResponseError(c, 2005, errors.New("user not found"))
		return
	}

	userTobeDeleted.IsDelete = 1

	err = userTobeDeleted.Save(c, db)
	if err != nil {
		middleware.ResponseWithCode(c, http.StatusInternalServerError, 2006, err, "db save action failed")
		return
	}

	middleware.ResponseSuccess(c, "用户删除成功！")
}

// UserMe godoc
// @Summary      获取我的信息（仅限管理员）
// @Description  获取我的信息
// @Tags         user
// @Produce      json
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /user/me [get]
func (u *UserLogoutController) UserMe(c *gin.Context) {
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
	user, err := search.GetMyInfo(c, db, sessInfo.Id)
	if err != nil {
		middleware.ResponseError(c, 2004, errors.New("an error occurred during fetching user profile"))
		return
	}

	out := dto.UserLoginOutput{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		Role:      user.Role,
	}

	middleware.ResponseSuccess(c, out)
}

// PrintGetRequest godoc
// @Summary      获取我的信息（仅限管理员）
// @Description  获取我的信息
// @Tags         user
// @Produce      json
// @Success      200  {object}  middleware.Response "success"
// @Failure      500  {object}  middleware.Response
// @Router       /user/debug/get [get]
func (u *UserLogoutController) PrintGetRequest(c *gin.Context) {
	out := fmt.Sprintf(`
<html>
	<p>RemoteIP: %s</p>
	<p>Request: %+v</p>
</html>`, c.RemoteIP(), c.Request)

	log.Println(out)

	c.String(200, out)
}
