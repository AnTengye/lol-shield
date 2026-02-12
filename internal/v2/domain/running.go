package domain

type RunningSnapshot struct {
	GameFlow  string                 `json:"gameFlow"`
	QueueID   int                    `json:"queueId"`
	QueueName string                 `json:"queueName"`
	SelfTeam  []RunningPlayer        `json:"selfTeam"`
	EnemyTeam []RunningPlayer        `json:"enemyTeam"`
	Groups    map[int][]string       `json:"groups"`
	SkinMap   map[string]SkinBinding `json:"skinMap"`
}

type SkinBinding struct {
	ChampionID int64 `json:"championId"`
	SkinID     int64 `json:"skinId"`
}

type RunningPlayer struct {
	Puuid      string          `json:"puuid"`
	SummonerID int64           `json:"summonerId"`
	GameName   string          `json:"gameName"`
	TagLine    string          `json:"tagLine"`
	History    []MatchHistory  `json:"history"`
	Highest    RankedBriefInfo `json:"highest"`
}

type RankedBriefInfo struct {
	Tier      string `json:"tier"`
	Division  string `json:"division"`
	QueueType string `json:"queueType"`
}

type MatchHistory struct {
	GameID     int64  `json:"gameId"`
	CreateTime int64  `json:"createTime"`
	QueueID    int64  `json:"queueId"`
	ChampionID int64  `json:"championId"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	Assists    int    `json:"assists"`
	Win        bool   `json:"win"`
	GameMode   string `json:"gameMode"`
}
