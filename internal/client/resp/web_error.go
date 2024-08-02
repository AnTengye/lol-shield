package resp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var Success = NewResp(ResponseCode{}, "OK")

// user
var (
	InputDataNotFound           = NewResp(NewResponseCode(UserOpr, 1), "输入数据不存在")
	CurrentDataConditionNotMeet = NewResp(NewResponseCode(UserOpr, 2), "当前对象状态不满足条件")
	InputDataFormatErr          = NewResp(NewResponseCode(UserOpr, 3), "入参格式错误")
	InputDataErr                = NewResp(NewResponseCode(UserOpr, 4), "入参数据错误")
)

// business
var (
	FileOperationError    = NewResp(NewResponseCode(BusinessLogic, 3), "文件操作失败")
	NotEnoughResources    = NewResp(NewResponseCode(BusinessLogic, 4), "没有足够资源")
	PermissionPolicyError = NewResp(NewResponseCode(BusinessLogic, 5), "权限数据异常")
	RelationDataError     = NewResp(NewResponseCode(BusinessLogic, 6), "关联数据异常")
	PermissionFailed      = NewResp(NewResponseCode(BusinessLogic, 7), "认证失败")
)

// third
var (
	LcuConnectErr = NewResp(NewResponseCode(ThirdPartyServices, 1), "客户端连接异常")
)

func WriteData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
func WriteList(c *gin.Context, total, data interface{}) {
	c.JSON(
		http.StatusOK, gin.H{
			"total":   total,
			"data":    data,
			"code":    0,
			"message": "OK",
		},
	)
}

func WriteRespData(c *gin.Context, data interface{}) {
	res := Success
	res.Data = data
	c.JSON(http.StatusOK, res)
}

func WriteErrRes(c *gin.Context, res *Resp) {
	c.JSON(500, res)
}
