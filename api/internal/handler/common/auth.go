package common

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/client"

	"rxcsoft.cn/pit3/api/internal/common/cryptox"
	"rxcsoft.cn/pit3/api/internal/common/httpx"
	"rxcsoft.cn/pit3/api/internal/common/loggerx"
	"rxcsoft.cn/pit3/api/internal/common/logic/ipx"
	"rxcsoft.cn/pit3/api/internal/middleware/auth/jwt"
	"rxcsoft.cn/pit3/api/internal/middleware/exist"
	"rxcsoft.cn/pit3/lib/msg"
	"rxcsoft.cn/pit3/srv/manage/proto/customer"
	"rxcsoft.cn/pit3/srv/manage/proto/role"
	"rxcsoft.cn/pit3/srv/manage/proto/user"
)

// Auth 认证
type Auth struct{}

// log出力
const (
	AuthProcessName    = "Auth"
	ActionLogin        = "Login"
	ActionRefreshToken = "RefreshToken"
)

// Login 登录
// @Router /login [post]
func (a *Auth) Login(c *gin.Context) {
	clientIP := c.ClientIP()
	loggerx.InfoLog(c, ActionLogin, loggerx.MsgProcessStarted)

	userService := user.NewUserService("manage", client.DefaultClient)

	var req user.LoginRequest
	// 从body中获取参数
	if err := c.BindJSON(&req); err != nil {
		httpx.GinHTTPError(c, ActionLogin, err)
		return
	}
	// 更换当前的密码为md5加密后的密码
	req.Password = cryptox.GenerateMd5Password(req.GetPassword(), req.GetEmail())
	email := req.GetEmail()

	var opss client.CallOption = func(o *client.CallOptions) {
		o.RequestTimeout = time.Second * 120
		o.DialTimeout = time.Second * 120
	}

	res, err := userService.Login(context.TODO(), &req, opss)
	if err != nil {
		loggerx.LoginLog(clientIP, email, "", ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, err), false)
		httpx.GinHTTPError(c, ActionLogin, err)
		return
	}

	if len(res.Error) > 0 {
		if res.Error == mongo.ErrNoDocuments.Error() {
			loggerx.LoginLog(clientIP, email, "", ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, err), false)
			c.JSON(200, httpx.Response{
				Status:  2,
				Message: msg.GetMsg("ja-JP", msg.Warn, msg.W001),
				Data:    gin.H{},
			})
			return
		}
		if res.Error == "password is invalid" {
			var req user.FindUserRequest
			req.Type = 1
			req.Email = email
			u, _ := userService.FindUser(context.TODO(), &req)
			loggerx.LoginLog(clientIP, email, u.GetUser().GetDomain(), ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, err), false)
			c.JSON(200, httpx.Response{
				Status:  2,
				Message: msg.GetMsg("ja-JP", msg.Warn, msg.W002, strconv.Itoa(int(u.GetUser().GetErrorCount()))),
				Data:    gin.H{},
			})
			return
		}
		if res.Error == "user has been locked" {
			var req user.FindUserRequest
			req.Type = 1
			req.Email = email
			u, _ := userService.FindUser(context.TODO(), &req)
			loggerx.LoginLog(clientIP, email, u.GetUser().GetDomain(), ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, err), false)
			c.JSON(200, httpx.Response{
				Status:  2,
				Message: msg.GetMsg("ja-JP", msg.Warn, msg.W005),
				Data:    gin.H{},
			})
			return
		}

		loggerx.LoginLog(clientIP, email, "", ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, res.Error), false)
		httpx.GinHTTPError(c, ActionLogin, errors.New(res.Error))
		return
	}

	// 判断用户登录IP是否设置白名单&判断用户登录IP是否在白名单中&判断角色类型
	var isUseIPSegment = false
	var inSegment = false
	var userFlg = 0
	if len(res.GetUser().GetRoles()) > 0 {
		for _, g := range res.GetUser().GetRoles() {
			loggerx.InfoLog(c, ActionLogin, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessStarted))
			roleService := role.NewRoleService("manage", client.DefaultClient)

			var req role.FindRoleRequest
			req.RoleId = g
			req.Database = res.GetUser().GetCustomerId()
			response, err := roleService.FindRole(context.TODO(), &req)
			if err != nil {
				httpx.GinHTTPError(c, ActionLogin, err)
				return
			}
			loggerx.InfoLog(c, ActionLogin, fmt.Sprintf("Process FindRole:%s", loggerx.MsgProcessEnded))

			// 用户登录IP是否设置白名单
			if len(response.Role.IpSegments) > 0 {
				isUseIPSegment = true
			}

			// 用户登录IP是否在白名单中
			if ipx.CheckIP(clientIP, response.Role.IpSegments) {
				inSegment = true
			}

			// 判断角色类型
			if response.Role.RoleType == 2 {
				// 超级管理员
				userFlg = 2
			}
			// 判断角色类型
			if response.Role.RoleType == 3 {
				// dev管理员
				userFlg = 3
			}
			if response.Role.RoleType == 1 {
				// 管理员
				userFlg = 1
			}

			// 判断是否已经合法
			if inSegment && userFlg != 0 {
				break
			}
		}
	}

	// 启用IP白名单并且用户登录IP不在白名单中
	if isUseIPSegment {
		if !inSegment {
			err := errors.New("the login ip is not in the ip whitelist")
			loggerx.LoginLog(clientIP, email, res.GetUser().GetDomain(), ActionLogin, fmt.Sprintf(loggerx.MsgProcessError, ActionLogin, err), false)
			c.JSON(200, httpx.Response{
				Status:  2,
				Message: msg.GetMsg("ja-JP", msg.Warn, msg.W004),
				Data: gin.H{
					"user_flg": userFlg,
				},
			})
			return
		}
	}

	loggerx.LoginLog(clientIP, email, res.GetUser().GetDomain(), ActionLogin, fmt.Sprintf(loggerx.MsgProcesSucceed, ActionLogin), true)

	// 判断用户登录当前APP情报是否有效
	var isValidApp = true
	if res.GetUser().GetCurrentApp() != "" && res.GetUser().GetCurrentApp() != "system" {
		if exist.CheckAppExpired(res.GetUser().GetCustomerId(), res.GetUser().GetCurrentApp()) {
			isValidApp = false
		}
	}

	// 判断用户登录是否需要二次验证
	var isSecondCheck = false
	var customerInfo customer.Customer
	if res.GetUser().GetUserType() != 2 && res.GetUser().GetUserType() != 3 {
		loggerx.InfoLog(c, ActionLogin, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessStarted))

		customerService := customer.NewCustomerService("manage", client.DefaultClient)

		var cReq customer.FindCustomerRequest
		cReq.CustomerId = res.GetUser().GetCustomerId()
		cResponse, err := customerService.FindCustomer(context.TODO(), &cReq)
		if err != nil {
			httpx.GinHTTPError(c, ActionLogin, err)
			return
		}
		loggerx.InfoLog(c, ActionLogin, fmt.Sprintf("Process FindCustomer:%s", loggerx.MsgProcessEnded))
		isSecondCheck = cResponse.GetCustomer().GetSecondCheck()
		customerInfo = *cResponse.GetCustomer()
	}

	j := jwt.NewJWT()

	claims := jwt.CustomClaims{
		UserID:     res.GetUser().GetUserId(),
		Email:      res.GetUser().GetEmail(),
		Domain:     res.GetUser().GetDomain(),
		CustomerID: res.GetUser().GetCustomerId(),
	}
	claims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
	claims.Issuer = "pit3"

	accessToken, err := j.CreateToken(claims)
	if err != nil {
		httpx.GinHTTPError(c, ActionLogin, err)
		return
	}

	claims.ExpiresAt = time.Now().Add(8 * time.Hour).Unix()
	refreshToken, err := j.CreateToken(claims)
	if err != nil {
		httpx.GinHTTPError(c, ActionLogin, err)
		return
	}

	type User struct {
		user.User
		Logo         string `json:"logo"`
		CustomerName string `json:"customer_name"`
	}

	userInfo := &User{
		*res.User,
		customerInfo.GetCustomerLogo(),
		customerInfo.GetCustomerName(),
	}

	loggerx.InfoLog(c, ActionLogin, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: msg.GetMsg("ja-JP", msg.Info, msg.I001),
		Data: gin.H{
			"access_token":    accessToken,
			"refresh_token":   refreshToken,
			"user":            userInfo,
			"user_flg":        userFlg,
			"is_valid_app":    isValidApp,
			"is_second_check": isSecondCheck,
		},
	})
}

// RefreshToken 刷新token
// @Router /refresh/token [post]
func (a *Auth) RefreshToken(c *gin.Context) {
	loggerx.InfoLog(c, ActionRefreshToken, loggerx.MsgProcessStarted)

	type RefreshReq struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req RefreshReq

	if err := c.BindJSON(&req); err != nil {
		httpx.GinTokenError(c, ActionRefreshToken, err)
		return
	}

	j := jwt.NewJWT()
	_, err := j.ParseToken(req.RefreshToken)
	if err != nil {
		httpx.GinTokenError(c, ActionRefreshToken, err)
		return
	}

	accessToken, e := j.RefreshToken(req.RefreshToken)
	if e != nil {
		httpx.GinTokenError(c, ActionRefreshToken, err)
		return
	}

	loggerx.InfoLog(c, ActionRefreshToken, loggerx.MsgProcessEnded)
	c.JSON(200, httpx.Response{
		Status:  0,
		Message: "重新刷新成功",
		Data: gin.H{
			"access_token": accessToken,
		},
	})
}
