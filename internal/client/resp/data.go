package resp

import (
	"fmt"
)

// Resp 返回结果
type Resp struct {
	// 错误编码，用于检索具体的错误类型
	//
	// * 0 成功
	// * 非0的错误码格式分为三个部分：A|05|023
	//
	// * A：错误行为.
	// ** A=用户行为-如用户的输入参数、用户的前置条件不满足
	// ** B=功能业务逻辑.
	// ** C=第三方服务
	// * 05:错误模块
	// * 023:具体的错误码
	Code ResponseCode `json:"code" swaggertype:"string"`
	// 错误消息
	//
	// * 成功        表示成功
	// * 其它     表示错误描述
	Message string `json:"message"`

	// 字段提示
	//
	// * 接口返回错误后的字段提示信息
	Field string `json:"field"`

	// 返回的数据对象
	//
	// * 具体的返回结果，`JSON`对象，具体接口会提供说明描述
	Data interface{} `json:"data,omitempty"`
}

func (e *Resp) WithField(f string) *Resp {
	e.Field = f
	return e
}

func (e *Resp) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

// NewResp returns a Resp
func NewResp(code ResponseCode, message string) *Resp {
	return &Resp{
		Code:    code,
		Message: message,
	}
}
