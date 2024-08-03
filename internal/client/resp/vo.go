package resp

type User struct {
	AccountId     int64  `json:"accountId"` //账户编号
	GameName      string `json:"gameName"`  //实际游戏昵称
	ProfileIconId int    `json:"profileIconId"`
	Level         int    `json:"summonerLevel"` //等级
	TagLine       string `json:"tagLine"`       //名称编号
	Tier          string `json:"tier"`
	Division      string `json:"division"`
	IsProvisional bool   `json:"isProvisional"` // 是否定位赛
}
