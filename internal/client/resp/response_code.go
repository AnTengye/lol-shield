package resp

import (
	"fmt"
)

type ErrorType string

const (
	SystemType         ErrorType = "Z" // 系统类型，仅限全局使用
	UserOpr            ErrorType = "A" // 用户行为-如用户的输入参数、用户的前置条件不满足
	BusinessLogic      ErrorType = "B" // 功能业务逻辑
	ThirdPartyServices ErrorType = "C" // 第三方服务
)

type ResponseCode struct {
	errorType  ErrorType
	customCode int
}

func (resCode *ResponseCode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + resCode.String() + `"`), nil
}

func (resCode ResponseCode) String() string {
	if resCode.customCode == 0 {
		return "0"
	}
	return fmt.Sprintf("%s%03d", resCode.errorType, resCode.customCode)
}

func NewResponseCode(errorType ErrorType, customCode int) ResponseCode {
	return ResponseCode{
		errorType:  errorType,
		customCode: customCode,
	}
}
