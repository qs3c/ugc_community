package web

import (
	"fmt"
	"gitee.com/geekbang/basic-go/webook/internal/service"
	"gitee.com/geekbang/basic-go/webook/internal/service/oauth2/wechat"
	ijwt "gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.UserService
	ijwt.Handler
	key             []byte
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service,
	hdl ijwt.Handler,
	userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgB"),
		stateCookieName: "jwt-state",
		Handler:         hdl,
	}
}

func (o *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", o.Auth2URL)
	g.Any("/callback", o.Callback)
}

func (o *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	// 产生段state然后调用服务根据state产生url
	// 并且用state产生jwt并存入cookie
	// 而后就以Result结构体的格式返回产生的url-第三方也就是微信的二维码页面
	// 当然这个Result是web中定义的和wechat中的Result不同
	state := uuid.New()
	val, err := o.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	err = o.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, Result{
		Data: val,
	})
}

func (o *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	// 验证刚刚存在cookie中的jwt里的state和url中的state是否相同
	err := o.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	// 你校验不校验都可以
	code := ctx.Query("code")
	// state := ctx.Query("state")
	wechatInfo, err := o.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}
	u, err := o.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	err = o.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
	return
}

// 把url中 ？后的参数state取出来 和 cookie 中的jwt解析出来的stateClaims中取出state做对比
// 所以把 jwt token 放在一个叫 jwt-state 的 cookie 中了？ 真的假的没有用 session 吗

func (o *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")
	ck, err := ctx.Cookie(o.stateCookieName)
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return o.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}

// 设置 stateCookie,填入state然后就拿这个claims去做jwt签名了，得到 tokenStr
// 根据state产生jwt签名并设置cookie

func (o *OAuth2WechatHandler) setStateCookie(ctx *gin.Context,
	state string) error {
	claims := StateClaims{
		State: state,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(o.key)
	if err != nil {

		return err
	}
	ctx.SetCookie(o.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil
}

// 组合了 jwt 包官方字段的 stateClaims

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
