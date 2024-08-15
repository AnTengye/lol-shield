package lcu

import (
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"go.uber.org/zap"
)

type (
	// 每单位的数据
	PerMinDeltas struct {
		Ten    float64 `json:"0-10"`
		Twenty float64 `json:"10-20"`
		Thirty float64 `json:"20-30"`
		Forty  float64 `json:"30-40"`
		Fifty  float64 `json:"40-50"`
		Sixty  float64 `json:"50-60"`
	}
)
type (
	ChampSelectPatchType string // 英雄选择会话更新类型
	ConversationMsgType  string // 会话组消息类型
)

type (
	CommonResp struct {
		ErrorCode  string `json:"errorCode"`
		HttpStatus int    `json:"httpStatus"`
		Message    string `json:"message"`
	}
	// 自己的数据
	CurrSummoner struct {
		AccountId                   int64  `json:"accountId"`    //账户编号
		DisplayName                 string `json:"displayName"`  //展示名称（目测是曾用名）
		GameName                    string `json:"gameName"`     //实际游戏昵称
		InternalName                string `json:"internalName"` //展示名称（目测是曾用名）
		NameChangeFlag              bool   `json:"nameChangeFlag"`
		PercentCompleteForNextLevel int    `json:"percentCompleteForNextLevel"`
		Privacy                     string `json:"privacy"`
		ProfileIconId               int    `json:"profileIconId"`
		Puuid                       string `json:"puuid"`
		RerollPoints                struct {
			CurrentPoints    int `json:"currentPoints"`
			MaxRolls         int `json:"maxRolls"`
			NumberOfRolls    int `json:"numberOfRolls"`
			PointsCostToRoll int `json:"pointsCostToRoll"`
			PointsToReroll   int `json:"pointsToReroll"`
		} `json:"rerollPoints"`
		SummonerId       int64  `json:"summonerId"`    //同accountId
		SummonerLevel    int    `json:"summonerLevel"` //等级
		TagLine          string `json:"tagLine"`       //名称编号
		Unnamed          bool   `json:"unnamed"`
		XpSinceLastLevel int    `json:"xpSinceLastLevel"`
		XpUntilNextLevel int    `json:"xpUntilNextLevel"`
	}
	GameListResp struct {
		CommonResp
		AccountID int64    `json:"accountId"`
		Games     GameList `json:"games"`
	}
	GameList struct {
		GameBeginDate  string     `json:"gameBeginDate"`
		GameCount      int        `json:"gameCount"`
		GameEndDate    string     `json:"gameEndDate"`
		GameIndexBegin int        `json:"gameIndexBegin"`
		GameIndexEnd   int        `json:"gameIndexEnd"`
		Games          []GameInfo `json:"games"`
	}
	GameInfo struct {
		GameCreation          int64           `json:"gameCreation"` // 创建时间戳 ms
		GameCreationDate      time.Time       `json:"gameCreationDate"`
		GameDuration          int             `json:"gameDuration"` // 游戏持续时长 秒
		GameId                int64           `json:"gameId"`
		GameMode              models.GameMode `json:"gameMode"`
		GameType              models.GameType `json:"gameType"`
		GameVersion           string          `json:"gameVersion"`
		MapId                 int             `json:"mapId"` // 地图id
		ParticipantIdentities []struct {      // 参与者
			ParticipantId int      `json:"participantId"` // 参与者id
			Player        struct { // 玩家信息
				AccountId         int    `json:"accountId"`
				CurrentAccountId  int    `json:"currentAccountId"`
				CurrentPlatformId string `json:"currentPlatformId"`
				GameName          string `json:"gameName"`
				MatchHistoryUri   string `json:"matchHistoryUri"`
				PlatformId        string `json:"platformId"`
				ProfileIcon       int    `json:"profileIcon"`
				Puuid             string `json:"puuid"`
				SummonerId        int64  `json:"summonerId"`
				SummonerName      string `json:"summonerName"`
				TagLine           string `json:"tagLine"`
			} `json:"player"`
		} `json:"participantIdentities"`
		Participants []struct { // 参与者详细信息
			ChampionId                models.Champion `json:"championId"` // 英雄id
			HighestAchievedSeasonTier string          `json:"highestAchievedSeasonTier"`
			ParticipantId             int             `json:"participantId"`
			Spell1Id                  models.Spell    `json:"spell1Id"` // 召唤师技能1
			Spell2Id                  models.Spell    `json:"spell2Id"` // 召唤师技能2
			Stats                     struct {
				Assists                         int  `json:"assists"`                   // 助攻数
				CausedEarlySurrender            bool `json:"causedEarlySurrender"`      // 是否申请了提前投降
				ChampLevel                      int  `json:"champLevel"`                // 召唤师等级
				CombatPlayerScore               int  `json:"combatPlayerScore"`         //
				DamageDealtToObjectives         int  `json:"damageDealtToObjectives"`   // 对战略点的总伤害
				DamageDealtToTurrets            int  `json:"damageDealtToTurrets"`      // 对防御塔的总伤害
				DamageSelfMitigated             int  `json:"damageSelfMitigated"`       // 自我缓和的生命值
				Deaths                          int  `json:"deaths"`                    // 死亡次数
				DoubleKills                     int  `json:"doubleKills"`               // 双杀次数
				EarlySurrenderAccomplice        bool `json:"earlySurrenderAccomplice"`  // 是否同意了提前投降
				FirstBloodAssist                bool `json:"firstBloodAssist"`          // 是否助攻了一血
				FirstBloodKill                  bool `json:"firstBloodKill"`            // 是否获得了一血
				FirstInhibitorAssist            bool `json:"firstInhibitorAssist"`      // 是否助攻了摧毁第一个水晶
				FirstInhibitorKill              bool `json:"firstInhibitorKill"`        // 是否摧毁了摧毁第一个水晶
				FirstTowerAssist                bool `json:"firstTowerAssist"`          // 是否助攻了摧毁一塔
				FirstTowerKill                  bool `json:"firstTowerKill"`            // 是否摧毁了一塔
				GameEndedInEarlySurrender       bool `json:"gameEndedInEarlySurrender"` // 游戏是否由提前投降结束的
				GameEndedInSurrender            bool `json:"gameEndedInSurrender"`      // 游戏是由投降结束的
				GoldEarned                      int  `json:"goldEarned"`                // 金币获取
				GoldSpent                       int  `json:"goldSpent"`                 // 金币使用
				InhibitorKills                  int  `json:"inhibitorKills"`            // 摧毁水晶数
				Item0                           int  `json:"item0"`                     // 物品1
				Item1                           int  `json:"item1"`
				Item2                           int  `json:"item2"`
				Item3                           int  `json:"item3"`
				Item4                           int  `json:"item4"`
				Item5                           int  `json:"item5"`
				Item6                           int  `json:"item6"`
				KillingSprees                   int  `json:"killingSprees"`                   // 多杀
				Kills                           int  `json:"kills"`                           // 击杀
				LargestCriticalStrike           int  `json:"largestCriticalStrike"`           // 最大暴击伤害
				LargestKillingSpree             int  `json:"largestKillingSpree"`             // 最高连杀
				LargestMultiKill                int  `json:"largestMultiKill"`                // 多杀次数
				LongestTimeSpentLiving          int  `json:"longestTimeSpentLiving"`          // 最长存活时间
				MagicDamageDealt                int  `json:"magicDamageDealt"`                // 造成的魔法伤害
				MagicDamageDealtToChampions     int  `json:"magicDamageDealtToChampions"`     // 对英雄造成的魔法伤害
				MagicalDamageTaken              int  `json:"magicalDamageTaken"`              // 承受的魔法伤害
				NeutralMinionsKilled            int  `json:"neutralMinionsKilled"`            // 击杀野怪
				NeutralMinionsKilledEnemyJungle int  `json:"neutralMinionsKilledEnemyJungle"` // 击杀敌方野怪
				NeutralMinionsKilledTeamJungle  int  `json:"neutralMinionsKilledTeamJungle"`  // 击杀队伍野怪
				ObjectivePlayerScore            int  `json:"objectivePlayerScore"`            //
				ParticipantId                   int  `json:"participantId"`
				PentaKills                      int  `json:"pentaKills"`
				Perk0                           int  `json:"perk0"`
				Perk0Var1                       int  `json:"perk0Var1"`
				Perk0Var2                       int  `json:"perk0Var2"`
				Perk0Var3                       int  `json:"perk0Var3"`
				Perk1                           int  `json:"perk1"`
				Perk1Var1                       int  `json:"perk1Var1"`
				Perk1Var2                       int  `json:"perk1Var2"`
				Perk1Var3                       int  `json:"perk1Var3"`
				Perk2                           int  `json:"perk2"`
				Perk2Var1                       int  `json:"perk2Var1"`
				Perk2Var2                       int  `json:"perk2Var2"`
				Perk2Var3                       int  `json:"perk2Var3"`
				Perk3                           int  `json:"perk3"`
				Perk3Var1                       int  `json:"perk3Var1"`
				Perk3Var2                       int  `json:"perk3Var2"`
				Perk3Var3                       int  `json:"perk3Var3"`
				Perk4                           int  `json:"perk4"`
				Perk4Var1                       int  `json:"perk4Var1"`
				Perk4Var2                       int  `json:"perk4Var2"`
				Perk4Var3                       int  `json:"perk4Var3"`
				Perk5                           int  `json:"perk5"`
				Perk5Var1                       int  `json:"perk5Var1"`
				Perk5Var2                       int  `json:"perk5Var2"`
				Perk5Var3                       int  `json:"perk5Var3"`
				PerkPrimaryStyle                int  `json:"perkPrimaryStyle"`
				PerkSubStyle                    int  `json:"perkSubStyle"`
				PhysicalDamageDealt             int  `json:"physicalDamageDealt"`            // 造成的物理伤害
				PhysicalDamageDealtToChampions  int  `json:"physicalDamageDealtToChampions"` // 对英雄造成的物理伤害
				PhysicalDamageTaken             int  `json:"physicalDamageTaken"`            // 受到的物理伤害
				PlayerScore0                    int  `json:"playerScore0"`
				PlayerScore1                    int  `json:"playerScore1"`
				PlayerScore2                    int  `json:"playerScore2"`
				PlayerScore3                    int  `json:"playerScore3"`
				PlayerScore4                    int  `json:"playerScore4"`
				PlayerScore5                    int  `json:"playerScore5"`
				PlayerScore6                    int  `json:"playerScore6"`
				PlayerScore7                    int  `json:"playerScore7"`
				PlayerScore8                    int  `json:"playerScore8"`
				PlayerScore9                    int  `json:"playerScore9"`
				QuadraKills                     int  `json:"quadraKills"`            // 四杀次数
				SightWardsBoughtInGame          int  `json:"sightWardsBoughtInGame"` //
				TeamEarlySurrendered            bool `json:"teamEarlySurrendered"`   // 队伍是否提前投降
				TimeCCingOthers                 int  `json:"timeCCingOthers"`
				TotalDamageDealt                int  `json:"totalDamageDealt"`            // 造成的伤害总和
				TotalDamageDealtToChampions     int  `json:"totalDamageDealtToChampions"` // 对英雄造成的伤害总和
				TotalDamageTaken                int  `json:"totalDamageTaken"`            // 对防御塔造成的伤害总和
				TotalHeal                       int  `json:"totalHeal"`                   // 治疗伤害
				TotalMinionsKilled              int  `json:"totalMinionsKilled"`          // 击杀小兵数
				TotalPlayerScore                int  `json:"totalPlayerScore"`
				TotalScoreRank                  int  `json:"totalScoreRank"`
				TotalTimeCrowdControlDealt      int  `json:"totalTimeCrowdControlDealt"` // 总控制时长
				TotalUnitsHealed                int  `json:"totalUnitsHealed"`           //
				TripleKills                     int  `json:"tripleKills"`                // 三杀次数
				TrueDamageDealt                 int  `json:"trueDamageDealt"`            //  总真实伤害
				TrueDamageDealtToChampions      int  `json:"trueDamageDealtToChampions"` // 对英雄的总真实伤害
				TrueDamageTaken                 int  `json:"trueDamageTaken"`            // 对防御塔的真实伤害
				TurretKills                     int  `json:"turretKills"`                // 击杀防御塔
				UnrealKills                     int  `json:"unrealKills"`                // 摧毁水晶
				VisionScore                     int  `json:"visionScore"`                // 视野得分
				VisionWardsBoughtInGame         int  `json:"visionWardsBoughtInGame"`    // 购买控制守卫
				WardsKilled                     int  `json:"wardsKilled"`                // 击杀守卫
				WardsPlaced                     int  `json:"wardsPlaced"`                // 放置守卫
				Win                             bool `json:"win"`                        // 是否获胜
			} `json:"stats"`
			TeamId   int `json:"teamId"`
			Timeline struct {
				CreepsPerMinDeltas          PerMinDeltas `json:"creepsPerMinDeltas"` // 每单位(分钟)移动码数(估计是千码)
				CsDiffPerMinDeltas          PerMinDeltas `json:"csDiffPerMinDeltas"`
				DamageTakenDiffPerMinDeltas PerMinDeltas `json:"damageTakenDiffPerMinDeltas"` // 每单位受到伤害差距
				DamageTakenPerMinDeltas     PerMinDeltas `json:"damageTakenPerMinDeltas"`     // 每单位受到伤害
				GoldPerMinDeltas            PerMinDeltas `json:"goldPerMinDeltas"`            // 每单位获得金币
				Lane                        string       `json:"lane"`                        // 哪一路
				ParticipantId               int          `json:"participantId"`               // 参与者id
				Role                        string       `json:"role"`                        // 角色
				XpDiffPerMinDeltas          PerMinDeltas `json:"xpDiffPerMinDeltas"`          // 每单位经验差距
				XpPerMinDeltas              PerMinDeltas `json:"xpPerMinDeltas"`              // 每单位经验数
			} `json:"timeline"`
		} `json:"participants"`
		PlatformId string             `json:"platformId"` // 平台id
		QueueId    models.GameQueueID `json:"queueId"`    // 队列id
		SeasonId   int                `json:"seasonId"`
		Teams      []interface{}      `json:"teams"`
	}
	// 聊天组
	Conversation struct {
		GameName           string          `json:"gameName"`
		GameTag            string          `json:"gameTag"`
		Id                 string          `json:"id"`
		InviterId          string          `json:"inviterId"`
		IsMuted            bool            `json:"isMuted"`
		LastMessage        interface{}     `json:"lastMessage"`
		Name               string          `json:"name"`
		Password           string          `json:"password"`
		Pid                string          `json:"pid"`
		TargetRegion       string          `json:"targetRegion"`
		Type               models.ChatType `json:"type"` // 聊天类型
		UnreadMessageCount int             `json:"unreadMessageCount"`
	}
	ConversationMsg struct {
		Body           string              `json:"body"`
		FromId         string              `json:"fromId"`
		FromPid        string              `json:"fromPid"`
		FromSummonerId int64               `json:"fromSummonerId"`
		Id             string              `json:"id"`
		IsHistorical   bool                `json:"isHistorical"`
		Timestamp      time.Time           `json:"timestamp"`
		Type           ConversationMsgType `json:"type"`
	}
	Summoner struct {
		CommonResp
		AccountId                   int64  `json:"accountId"`
		DisplayName                 string `json:"displayName"`
		InternalName                string `json:"internalName"`
		NameChangeFlag              bool   `json:"nameChangeFlag"`
		PercentCompleteForNextLevel int    `json:"percentCompleteForNextLevel"`
		Privacy                     string `json:"privacy"`
		ProfileIconId               int    `json:"profileIconId"`
		Puuid                       string `json:"puuid"`
		RerollPoints                struct {
			CurrentPoints    int `json:"currentPoints"`
			MaxRolls         int `json:"maxRolls"`
			NumberOfRolls    int `json:"numberOfRolls"`
			PointsCostToRoll int `json:"pointsCostToRoll"`
			PointsToReroll   int `json:"pointsToReroll"`
		} `json:"rerollPoints"`
		SummonerId       int64 `json:"summonerId"`
		SummonerLevel    int   `json:"summonerLevel"`
		Unnamed          bool  `json:"unnamed"`
		XpSinceLastLevel int   `json:"xpSinceLastLevel"`
		XpUntilNextLevel int   `json:"xpUntilNextLevel"`
	}
	Participant struct {
		ChampionId                int    `json:"championId"`
		HighestAchievedSeasonTier string `json:"highestAchievedSeasonTier"`
		ParticipantId             int    `json:"participantId"`
		Spell1Id                  int    `json:"spell1Id"`
		Spell2Id                  int    `json:"spell2Id"`
		Stats                     struct {
			Assists                         int  `json:"assists"`
			CausedEarlySurrender            bool `json:"causedEarlySurrender"`
			ChampLevel                      int  `json:"champLevel"`
			CombatPlayerScore               int  `json:"combatPlayerScore"`
			DamageDealtToObjectives         int  `json:"damageDealtToObjectives"`
			DamageDealtToTurrets            int  `json:"damageDealtToTurrets"`
			DamageSelfMitigated             int  `json:"damageSelfMitigated"`
			Deaths                          int  `json:"deaths"`
			DoubleKills                     int  `json:"doubleKills"`
			EarlySurrenderAccomplice        bool `json:"earlySurrenderAccomplice"`
			FirstBloodAssist                bool `json:"firstBloodAssist"`
			FirstBloodKill                  bool `json:"firstBloodKill"`
			FirstInhibitorAssist            bool `json:"firstInhibitorAssist"`
			FirstInhibitorKill              bool `json:"firstInhibitorKill"`
			FirstTowerAssist                bool `json:"firstTowerAssist"`
			FirstTowerKill                  bool `json:"firstTowerKill"`
			GameEndedInEarlySurrender       bool `json:"gameEndedInEarlySurrender"`
			GameEndedInSurrender            bool `json:"gameEndedInSurrender"`
			GoldEarned                      int  `json:"goldEarned"`
			GoldSpent                       int  `json:"goldSpent"`
			InhibitorKills                  int  `json:"inhibitorKills"`
			Item0                           int  `json:"item0"`
			Item1                           int  `json:"item1"`
			Item2                           int  `json:"item2"`
			Item3                           int  `json:"item3"`
			Item4                           int  `json:"item4"`
			Item5                           int  `json:"item5"`
			Item6                           int  `json:"item6"`
			KillingSprees                   int  `json:"killingSprees"`
			Kills                           int  `json:"kills"`
			LargestCriticalStrike           int  `json:"largestCriticalStrike"`
			LargestKillingSpree             int  `json:"largestKillingSpree"`
			LargestMultiKill                int  `json:"largestMultiKill"`
			LongestTimeSpentLiving          int  `json:"longestTimeSpentLiving"`
			MagicDamageDealt                int  `json:"magicDamageDealt"`
			MagicDamageDealtToChampions     int  `json:"magicDamageDealtToChampions"`
			MagicalDamageTaken              int  `json:"magicalDamageTaken"`
			NeutralMinionsKilled            int  `json:"neutralMinionsKilled"`
			NeutralMinionsKilledEnemyJungle int  `json:"neutralMinionsKilledEnemyJungle"`
			NeutralMinionsKilledTeamJungle  int  `json:"neutralMinionsKilledTeamJungle"`
			ObjectivePlayerScore            int  `json:"objectivePlayerScore"`
			ParticipantId                   int  `json:"participantId"`
			PentaKills                      int  `json:"pentaKills"`
			Perk0                           int  `json:"perk0"`
			Perk0Var1                       int  `json:"perk0Var1"`
			Perk0Var2                       int  `json:"perk0Var2"`
			Perk0Var3                       int  `json:"perk0Var3"`
			Perk1                           int  `json:"perk1"`
			Perk1Var1                       int  `json:"perk1Var1"`
			Perk1Var2                       int  `json:"perk1Var2"`
			Perk1Var3                       int  `json:"perk1Var3"`
			Perk2                           int  `json:"perk2"`
			Perk2Var1                       int  `json:"perk2Var1"`
			Perk2Var2                       int  `json:"perk2Var2"`
			Perk2Var3                       int  `json:"perk2Var3"`
			Perk3                           int  `json:"perk3"`
			Perk3Var1                       int  `json:"perk3Var1"`
			Perk3Var2                       int  `json:"perk3Var2"`
			Perk3Var3                       int  `json:"perk3Var3"`
			Perk4                           int  `json:"perk4"`
			Perk4Var1                       int  `json:"perk4Var1"`
			Perk4Var2                       int  `json:"perk4Var2"`
			Perk4Var3                       int  `json:"perk4Var3"`
			Perk5                           int  `json:"perk5"`
			Perk5Var1                       int  `json:"perk5Var1"`
			Perk5Var2                       int  `json:"perk5Var2"`
			Perk5Var3                       int  `json:"perk5Var3"`
			PerkPrimaryStyle                int  `json:"perkPrimaryStyle"`
			PerkSubStyle                    int  `json:"perkSubStyle"`
			PhysicalDamageDealt             int  `json:"physicalDamageDealt"`
			PhysicalDamageDealtToChampions  int  `json:"physicalDamageDealtToChampions"`
			PhysicalDamageTaken             int  `json:"physicalDamageTaken"`
			PlayerScore0                    int  `json:"playerScore0"`
			PlayerScore1                    int  `json:"playerScore1"`
			PlayerScore2                    int  `json:"playerScore2"`
			PlayerScore3                    int  `json:"playerScore3"`
			PlayerScore4                    int  `json:"playerScore4"`
			PlayerScore5                    int  `json:"playerScore5"`
			PlayerScore6                    int  `json:"playerScore6"`
			PlayerScore7                    int  `json:"playerScore7"`
			PlayerScore8                    int  `json:"playerScore8"`
			PlayerScore9                    int  `json:"playerScore9"`
			QuadraKills                     int  `json:"quadraKills"`
			SightWardsBoughtInGame          int  `json:"sightWardsBoughtInGame"`
			TeamEarlySurrendered            bool `json:"teamEarlySurrendered"`
			TimeCCingOthers                 int  `json:"timeCCingOthers"`
			TotalDamageDealt                int  `json:"totalDamageDealt"`
			TotalDamageDealtToChampions     int  `json:"totalDamageDealtToChampions"`
			TotalDamageTaken                int  `json:"totalDamageTaken"`
			TotalHeal                       int  `json:"totalHeal"`
			TotalMinionsKilled              int  `json:"totalMinionsKilled"`
			TotalPlayerScore                int  `json:"totalPlayerScore"`
			TotalScoreRank                  int  `json:"totalScoreRank"`
			TotalTimeCrowdControlDealt      int  `json:"totalTimeCrowdControlDealt"`
			TotalUnitsHealed                int  `json:"totalUnitsHealed"`
			TripleKills                     int  `json:"tripleKills"`
			TrueDamageDealt                 int  `json:"trueDamageDealt"`
			TrueDamageDealtToChampions      int  `json:"trueDamageDealtToChampions"`
			TrueDamageTaken                 int  `json:"trueDamageTaken"`
			TurretKills                     int  `json:"turretKills"`
			UnrealKills                     int  `json:"unrealKills"`
			VisionScore                     int  `json:"visionScore"`
			VisionWardsBoughtInGame         int  `json:"visionWardsBoughtInGame"`
			WardsKilled                     int  `json:"wardsKilled"`
			WardsPlaced                     int  `json:"wardsPlaced"`
			Win                             bool `json:"win"`
		} `json:"stats"`
		TeamId   models.TeamID `json:"teamId"`
		Timeline struct {
			CreepsPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"creepsPerMinDeltas"`
			CsDiffPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"csDiffPerMinDeltas"`
			DamageTakenDiffPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"damageTakenDiffPerMinDeltas"`
			DamageTakenPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"damageTakenPerMinDeltas"`
			GoldPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"goldPerMinDeltas"`
			Lane               models.Lane         `json:"lane"`
			ParticipantId      int                 `json:"participantId"`
			Role               models.ChampionRole `json:"role"`
			XpDiffPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"xpDiffPerMinDeltas"`
			XpPerMinDeltas struct {
				Field1 float64 `json:"0-10"`
				Field2 float64 `json:"10-20"`
			} `json:"xpPerMinDeltas"`
		} `json:"timeline"`
	}
	ChampSelectSessionInfo struct {
		CommonResp
		Actions [][]struct {
			ActorCellId  int                  `json:"actorCellId"`
			ChampionId   int                  `json:"championId"`
			Completed    bool                 `json:"completed"`
			Id           int                  `json:"id"`
			IsAllyAction bool                 `json:"isAllyAction"`
			IsInProgress bool                 `json:"isInProgress"`
			PickTurn     int                  `json:"pickTurn"`
			Type         ChampSelectPatchType `json:"type"`
		} `json:"actions"`
		// AllowBattleBoost    bool `json:"allowBattleBoost"`
		// AllowDuplicatePicks bool `json:"allowDuplicatePicks"`
		// AllowLockedEvents   bool `json:"allowLockedEvents"`
		// AllowRerolling      bool `json:"allowRerolling"`
		// AllowSkinSelection  bool `json:"allowSkinSelection"`
		// Bans                struct {
		// 	MyTeamBans    []interface{} `json:"myTeamBans"`
		// 	NumBans       int           `json:"numBans"`
		// 	TheirTeamBans []interface{} `json:"theirTeamBans"`
		// } `json:"bans"`
		// BenchChampionIds   []interface{} `json:"benchChampionIds"`
		// BenchEnabled       bool          `json:"benchEnabled"`
		// BoostableSkinCount int           `json:"boostableSkinCount"`
		// ChatDetails        struct {
		// 	ChatRoomName     string `json:"chatRoomName"`
		// 	ChatRoomPassword string `json:"chatRoomPassword"`
		// } `json:"chatDetails"`
		// Counter              int `json:"counter"`
		// EntitledFeatureState struct {
		// 	AdditionalRerolls int           `json:"additionalRerolls"`
		// 	UnlockedSkinIds   []interface{} `json:"unlockedSkinIds"`
		// } `json:"entitledFeatureState"`
		// GameId               int  `json:"gameId"`
		// HasSimultaneousBans  bool `json:"hasSimultaneousBans"`
		// HasSimultaneousPicks bool `json:"hasSimultaneousPicks"`
		// IsCustomGame         bool `json:"isCustomGame"`
		// IsSpectating         bool `json:"isSpectating"`
		LocalPlayerCellId int `json:"localPlayerCellId"`
		// LockedEventIndex     int  `json:"lockedEventIndex"`
		// MyTeam               []struct {
		// 	AssignedPosition    string `json:"assignedPosition"`
		// 	CellId              int    `json:"cellId"`
		// 	ChampionId          int    `json:"championId"`
		// 	ChampionPickIntent  int    `json:"championPickIntent"`
		// 	EntitledFeatureType string `json:"entitledFeatureType"`
		// 	SelectedSkinId      int    `json:"selectedSkinId"`
		// 	Spell1Id            int    `json:"spell1Id"`
		// 	Spell2Id            int    `json:"spell2Id"`
		// 	SummonerId          int64  `json:"summonerId"`
		// 	Team                int    `json:"team"`
		// 	WardSkinId          int    `json:"wardSkinId"`
		// } `json:"myTeam"`
		// RecoveryCounter    int  `json:"recoveryCounter"`
		// RerollsRemaining   int  `json:"rerollsRemaining"`
		// SkipChampionSelect bool `json:"skipChampionSelect"`
		// TheirTeam          []struct {
		// 	AssignedPosition    string `json:"assignedPosition"`
		// 	CellId              int    `json:"cellId"`
		// 	ChampionId          int    `json:"championId"`
		// 	ChampionPickIntent  int    `json:"championPickIntent"`
		// 	EntitledFeatureType string `json:"entitledFeatureType"`
		// 	SelectedSkinId      int    `json:"selectedSkinId"`
		// 	Spell1Id            int    `json:"spell1Id"`
		// 	Spell2Id            int    `json:"spell2Id"`
		// 	SummonerId          int    `json:"summonerId"`
		// 	Team                int    `json:"team"`
		// 	WardSkinId          int    `json:"wardSkinId"`
		// } `json:"theirTeam"`
		// Timer struct {
		// 	AdjustedTimeLeftInPhase int    `json:"adjustedTimeLeftInPhase"`
		// 	InternalNowInEpochMs    int64  `json:"internalNowInEpochMs"`
		// 	IsInfinite              bool   `json:"isInfinite"`
		// 	Phase                   string `json:"phase"`
		// 	TotalTimeInPhase        int    `json:"totalTimeInPhase"`
		// } `json:"timer"`
		// Trades []interface{} `json:"trades"`
	}
)

type GameFlowSession struct {
	CommonResp
	GameClient struct {
		ObserverServerIp   string `json:"observerServerIp"`
		ObserverServerPort int    `json:"observerServerPort"`
		Running            bool   `json:"running"`
		ServerIp           string `json:"serverIp"`
		ServerPort         int    `json:"serverPort"`
		Visible            bool   `json:"visible"`
	} `json:"gameClient"`
	GameData struct {
		GameId                   int64  `json:"gameId"`
		GameName                 string `json:"gameName"`
		IsCustomGame             bool   `json:"isCustomGame"`
		Password                 string `json:"password"`
		PlayerChampionSelections []struct {
			ChampionId           int    `json:"championId"`
			SelectedSkinIndex    int    `json:"selectedSkinIndex"`
			Spell1Id             int    `json:"spell1Id"`
			Spell2Id             int    `json:"spell2Id"`
			SummonerInternalName string `json:"summonerInternalName"`
		} `json:"playerChampionSelections"`
		Queue struct {
			AllowablePremadeSizes   []int  `json:"allowablePremadeSizes"`
			AreFreeChampionsAllowed bool   `json:"areFreeChampionsAllowed"`
			AssetMutator            string `json:"assetMutator"`
			Category                string `json:"category"`
			ChampionsRequiredToPlay int    `json:"championsRequiredToPlay"`
			Description             string `json:"description"`
			DetailedDescription     string `json:"detailedDescription"`
			GameMode                string `json:"gameMode"`
			GameTypeConfig          struct {
				AdvancedLearningQuests bool   `json:"advancedLearningQuests"`
				AllowTrades            bool   `json:"allowTrades"`
				BanMode                string `json:"banMode"`
				BanTimerDuration       int    `json:"banTimerDuration"`
				BattleBoost            bool   `json:"battleBoost"`
				CrossTeamChampionPool  bool   `json:"crossTeamChampionPool"`
				DeathMatch             bool   `json:"deathMatch"`
				DoNotRemove            bool   `json:"doNotRemove"`
				DuplicatePick          bool   `json:"duplicatePick"`
				ExclusivePick          bool   `json:"exclusivePick"`
				Id                     int    `json:"id"`
				LearningQuests         bool   `json:"learningQuests"`
				MainPickTimerDuration  int    `json:"mainPickTimerDuration"`
				MaxAllowableBans       int    `json:"maxAllowableBans"`
				Name                   string `json:"name"`
				OnboardCoopBeginner    bool   `json:"onboardCoopBeginner"`
				PickMode               string `json:"pickMode"`
				PostPickTimerDuration  int    `json:"postPickTimerDuration"`
				Reroll                 bool   `json:"reroll"`
				TeamChampionPool       bool   `json:"teamChampionPool"`
			} `json:"gameTypeConfig"`
			Id                         int    `json:"id"`
			IsRanked                   bool   `json:"isRanked"`
			IsTeamBuilderManaged       bool   `json:"isTeamBuilderManaged"`
			LastToggledOffTime         int    `json:"lastToggledOffTime"`
			LastToggledOnTime          int    `json:"lastToggledOnTime"`
			MapId                      int    `json:"mapId"`
			MaximumParticipantListSize int    `json:"maximumParticipantListSize"`
			MinLevel                   int    `json:"minLevel"`
			MinimumParticipantListSize int    `json:"minimumParticipantListSize"`
			Name                       string `json:"name"`
			NumPlayersPerTeam          int    `json:"numPlayersPerTeam"`
			QueueAvailability          string `json:"queueAvailability"`
			QueueRewards               struct {
				IsChampionPointsEnabled bool          `json:"isChampionPointsEnabled"`
				IsIpEnabled             bool          `json:"isIpEnabled"`
				IsXpEnabled             bool          `json:"isXpEnabled"`
				PartySizeIpRewards      []interface{} `json:"partySizeIpRewards"`
			} `json:"queueRewards"`
			RemovalFromGameAllowed      bool   `json:"removalFromGameAllowed"`
			RemovalFromGameDelayMinutes int    `json:"removalFromGameDelayMinutes"`
			ShortName                   string `json:"shortName"`
			ShowPositionSelector        bool   `json:"showPositionSelector"`
			SpectatorEnabled            bool   `json:"spectatorEnabled"`
			Type                        string `json:"type"`
		} `json:"queue"`
		SpectatorsAllowed bool `json:"spectatorsAllowed"`
		TeamOne           []struct {
			ChampionId            int    `json:"championId"`
			LastSelectedSkinIndex int    `json:"lastSelectedSkinIndex"`
			ProfileIconId         int    `json:"profileIconId"`
			Puuid                 string `json:"puuid"`
			SelectedPosition      string `json:"selectedPosition"`
			SelectedRole          string `json:"selectedRole"`
			SummonerId            int64  `json:"summonerId"`
			SummonerInternalName  string `json:"summonerInternalName"`
			SummonerName          string `json:"summonerName"`
			TeamOwner             bool   `json:"teamOwner"`
			TeamParticipantId     int    `json:"teamParticipantId"`
		} `json:"teamOne"`
		TeamTwo []struct {
			ChampionId            int    `json:"championId"`
			LastSelectedSkinIndex int    `json:"lastSelectedSkinIndex"`
			ProfileIconId         int    `json:"profileIconId"`
			Puuid                 string `json:"puuid"`
			SelectedPosition      string `json:"selectedPosition"`
			SelectedRole          string `json:"selectedRole"`
			SummonerId            int64  `json:"summonerId"`
			SummonerInternalName  string `json:"summonerInternalName"`
			SummonerName          string `json:"summonerName"`
			TeamOwner             bool   `json:"teamOwner"`
			TeamParticipantId     int    `json:"teamParticipantId"`
		} `json:"teamTwo"`
	} `json:"gameData"`
	GameDodge struct {
		DodgeIds []interface{} `json:"dodgeIds"`
		Phase    string        `json:"phase"`
		State    string        `json:"state"`
	} `json:"gameDodge"`
	Map struct {
		Assets struct {
			ChampSelectBackgroundSound         string `json:"champ-select-background-sound"`
			ChampSelectBanphaseBackgroundSound string `json:"champ-select-banphase-background-sound"`
			ChampSelectFlyoutBackground        string `json:"champ-select-flyout-background"`
			GameSelectIconActive               string `json:"game-select-icon-active"`
			GameSelectIconActiveVideo          string `json:"game-select-icon-active-video"`
			GameSelectIconDefault              string `json:"game-select-icon-default"`
			GameSelectIconDisabled             string `json:"game-select-icon-disabled"`
			GameSelectIconHover                string `json:"game-select-icon-hover"`
			GameSelectIconIntroVideo           string `json:"game-select-icon-intro-video"`
			GameflowBackground                 string `json:"gameflow-background"`
			GameflowBackgroundDark             string `json:"gameflow-background-dark"`
			GameselectButtonHoverSound         string `json:"gameselect-button-hover-sound"`
			IconDefeat                         string `json:"icon-defeat"`
			IconDefeatV2                       string `json:"icon-defeat-v2"`
			IconDefeatVideo                    string `json:"icon-defeat-video"`
			IconEmpty                          string `json:"icon-empty"`
			IconHover                          string `json:"icon-hover"`
			IconLeaver                         string `json:"icon-leaver"`
			IconLeaverV2                       string `json:"icon-leaver-v2"`
			IconLossForgivenV2                 string `json:"icon-loss-forgiven-v2"`
			IconV2                             string `json:"icon-v2"`
			IconVictory                        string `json:"icon-victory"`
			IconVictoryVideo                   string `json:"icon-victory-video"`
			MusicInqueueLoopSound              string `json:"music-inqueue-loop-sound"`
			PartiesBackground                  string `json:"parties-background"`
			PostgameAmbienceLoopSound          string `json:"postgame-ambience-loop-sound"`
			ReadyCheckBackground               string `json:"ready-check-background"`
			ReadyCheckBackgroundSound          string `json:"ready-check-background-sound"`
			SfxAmbiencePregameLoopSound        string `json:"sfx-ambience-pregame-loop-sound"`
			SocialIconLeaver                   string `json:"social-icon-leaver"`
			SocialIconVictory                  string `json:"social-icon-victory"`
		} `json:"assets"`
		CategorizedContentBundles struct {
		} `json:"categorizedContentBundles"`
		Description                         string `json:"description"`
		GameMode                            string `json:"gameMode"`
		GameModeName                        string `json:"gameModeName"`
		GameModeShortName                   string `json:"gameModeShortName"`
		GameMutator                         string `json:"gameMutator"`
		Id                                  int    `json:"id"`
		IsRGM                               bool   `json:"isRGM"`
		MapStringId                         string `json:"mapStringId"`
		Name                                string `json:"name"`
		PerPositionDisallowedSummonerSpells struct {
		} `json:"perPositionDisallowedSummonerSpells"`
		PerPositionRequiredSummonerSpells struct {
		} `json:"perPositionRequiredSummonerSpells"`
		PlatformId   string `json:"platformId"`
		PlatformName string `json:"platformName"`
		Properties   struct {
			SuppressRunesMasteriesPerks bool `json:"suppressRunesMasteriesPerks"`
		} `json:"properties"`
	} `json:"map"`
	Phase models.GameStatus `json:"phase"`
}

type RankedData struct {
	CurrentSeasonSplitPoints          int           `json:"currentSeasonSplitPoints"`
	EarnedRegaliaRewardIds            []interface{} `json:"earnedRegaliaRewardIds"`
	HighestCurrentSeasonReachedTierSR string        `json:"highestCurrentSeasonReachedTierSR"`
	HighestPreviousSeasonEndDivision  string        `json:"highestPreviousSeasonEndDivision"`
	HighestPreviousSeasonEndTier      string        `json:"highestPreviousSeasonEndTier"`
	HighestRankedEntry                struct {
		Division                      string      `json:"division"`
		HighestDivision               string      `json:"highestDivision"`
		HighestTier                   string      `json:"highestTier"`
		IsProvisional                 bool        `json:"isProvisional"`
		LeaguePoints                  int         `json:"leaguePoints"`
		Losses                        int         `json:"losses"`
		MiniSeriesProgress            string      `json:"miniSeriesProgress"`
		PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
		PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
		PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
		PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
		ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
		ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
		QueueType                     string      `json:"queueType"`
		RatedRating                   int         `json:"ratedRating"`
		RatedTier                     string      `json:"ratedTier"`
		Tier                          string      `json:"tier"`
		Warnings                      interface{} `json:"warnings"`
		Wins                          int         `json:"wins"`
	} `json:"highestRankedEntry"`
	HighestRankedEntrySR struct {
		Division                      string      `json:"division"`
		HighestDivision               string      `json:"highestDivision"`
		HighestTier                   string      `json:"highestTier"`
		IsProvisional                 bool        `json:"isProvisional"`
		LeaguePoints                  int         `json:"leaguePoints"`
		Losses                        int         `json:"losses"`
		MiniSeriesProgress            string      `json:"miniSeriesProgress"`
		PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
		PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
		PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
		PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
		ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
		ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
		QueueType                     string      `json:"queueType"`
		RatedRating                   int         `json:"ratedRating"`
		RatedTier                     string      `json:"ratedTier"`
		Tier                          string      `json:"tier"`
		Warnings                      interface{} `json:"warnings"`
		Wins                          int         `json:"wins"`
	} `json:"highestRankedEntrySR"`
	PreviousSeasonSplitPoints int `json:"previousSeasonSplitPoints"`
	QueueMap                  struct {
		CHERRY struct {
			Division                      string      `json:"division"`
			HighestDivision               string      `json:"highestDivision"`
			HighestTier                   string      `json:"highestTier"`
			IsProvisional                 bool        `json:"isProvisional"`
			LeaguePoints                  int         `json:"leaguePoints"`
			Losses                        int         `json:"losses"`
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"`
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"`
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"`
		} `json:"CHERRY"`
		RANKEDFLEXSR struct {
			Division                      string      `json:"division"`
			HighestDivision               string      `json:"highestDivision"`
			HighestTier                   string      `json:"highestTier"`
			IsProvisional                 bool        `json:"isProvisional"`
			LeaguePoints                  int         `json:"leaguePoints"`
			Losses                        int         `json:"losses"`
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"`
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"`
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"`
		} `json:"RANKED_FLEX_SR"`
		RANKEDSOLO5X5 struct {
			Division                      string      `json:"division"`        //段位级别
			HighestDivision               string      `json:"highestDivision"` //最高段位级别
			HighestTier                   string      `json:"highestTier"`     //最高段位
			IsProvisional                 bool        `json:"isProvisional"`   //是否定位赛
			LeaguePoints                  int         `json:"leaguePoints"`    //分数
			Losses                        int         `json:"losses"`          //输的盘数
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"` //排位类型
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"` //当前段位
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"` //赢的盘数
		} `json:"RANKED_SOLO_5x5"`
		RANKEDTFT struct {
			Division                      string      `json:"division"`
			HighestDivision               string      `json:"highestDivision"`
			HighestTier                   string      `json:"highestTier"`
			IsProvisional                 bool        `json:"isProvisional"`
			LeaguePoints                  int         `json:"leaguePoints"`
			Losses                        int         `json:"losses"`
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"`
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"`
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"`
		} `json:"RANKED_TFT"`
		RANKEDTFTDOUBLEUP struct {
			Division                      string      `json:"division"`
			HighestDivision               string      `json:"highestDivision"`
			HighestTier                   string      `json:"highestTier"`
			IsProvisional                 bool        `json:"isProvisional"`
			LeaguePoints                  int         `json:"leaguePoints"`
			Losses                        int         `json:"losses"`
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"`
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"`
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"`
		} `json:"RANKED_TFT_DOUBLE_UP"`
		RANKEDTFTTURBO struct {
			Division                      string      `json:"division"`
			HighestDivision               string      `json:"highestDivision"`
			HighestTier                   string      `json:"highestTier"`
			IsProvisional                 bool        `json:"isProvisional"`
			LeaguePoints                  int         `json:"leaguePoints"`
			Losses                        int         `json:"losses"`
			MiniSeriesProgress            string      `json:"miniSeriesProgress"`
			PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
			PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
			PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
			PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
			ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
			ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
			QueueType                     string      `json:"queueType"`
			RatedRating                   int         `json:"ratedRating"`
			RatedTier                     string      `json:"ratedTier"`
			Tier                          string      `json:"tier"`
			Warnings                      interface{} `json:"warnings"`
			Wins                          int         `json:"wins"`
		} `json:"RANKED_TFT_TURBO"`
	} `json:"queueMap"`
	Queues []struct {
		Division                      string      `json:"division"`
		HighestDivision               string      `json:"highestDivision"`
		HighestTier                   string      `json:"highestTier"`
		IsProvisional                 bool        `json:"isProvisional"`
		LeaguePoints                  int         `json:"leaguePoints"`
		Losses                        int         `json:"losses"`
		MiniSeriesProgress            string      `json:"miniSeriesProgress"`
		PreviousSeasonEndDivision     string      `json:"previousSeasonEndDivision"`
		PreviousSeasonEndTier         string      `json:"previousSeasonEndTier"`
		PreviousSeasonHighestDivision string      `json:"previousSeasonHighestDivision"`
		PreviousSeasonHighestTier     string      `json:"previousSeasonHighestTier"`
		ProvisionalGameThreshold      int         `json:"provisionalGameThreshold"`
		ProvisionalGamesRemaining     int         `json:"provisionalGamesRemaining"`
		QueueType                     string      `json:"queueType"`
		RatedRating                   int         `json:"ratedRating"`
		RatedTier                     string      `json:"ratedTier"`
		Tier                          string      `json:"tier"`
		Warnings                      interface{} `json:"warnings"`
		Wins                          int         `json:"wins"`
	} `json:"queues"`
	RankedRegaliaLevel int `json:"rankedRegaliaLevel"`
	Seasons            struct {
		CHERRY struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"CHERRY"`
		RANKEDFLEXSR struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"RANKED_FLEX_SR"`
		RANKEDSOLO5X5 struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"RANKED_SOLO_5x5"`
		RANKEDTFT struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"RANKED_TFT"`
		RANKEDTFTDOUBLEUP struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"RANKED_TFT_DOUBLE_UP"`
		RANKEDTFTTURBO struct {
			CurrentSeasonEnd int64 `json:"currentSeasonEnd"`
			CurrentSeasonId  int   `json:"currentSeasonId"`
			NextSeasonStart  int   `json:"nextSeasonStart"`
		} `json:"RANKED_TFT_TURBO"`
	} `json:"seasons"`
	SplitsProgress struct {
		Field1 int `json:"2"`
	} `json:"splitsProgress"`
}

type GameSummary struct {
	CommonResp
	EndOfGameResult       string    `json:"endOfGameResult"`
	GameCreation          int64     `json:"gameCreation"`
	GameCreationDate      time.Time `json:"gameCreationDate"`
	GameDuration          int       `json:"gameDuration"`
	GameId                int64     `json:"gameId"`
	GameMode              string    `json:"gameMode"`
	GameType              string    `json:"gameType"`
	GameVersion           string    `json:"gameVersion"`
	MapId                 int       `json:"mapId"`
	ParticipantIdentities []struct {
		ParticipantId int `json:"participantId"`
		Player        struct {
			AccountId         int    `json:"accountId"`
			CurrentAccountId  int    `json:"currentAccountId"`
			CurrentPlatformId string `json:"currentPlatformId"`
			GameName          string `json:"gameName"`
			MatchHistoryUri   string `json:"matchHistoryUri"`
			PlatformId        string `json:"platformId"`
			ProfileIcon       int    `json:"profileIcon"`
			Puuid             string `json:"puuid"`
			SummonerId        int64  `json:"summonerId"`
			SummonerName      string `json:"summonerName"`
			TagLine           string `json:"tagLine"`
		} `json:"player"`
	} `json:"participantIdentities"`
	Participants []struct {
		ChampionId                int    `json:"championId"`
		HighestAchievedSeasonTier string `json:"highestAchievedSeasonTier"`
		ParticipantId             int    `json:"participantId"`
		Spell1Id                  int    `json:"spell1Id"`
		Spell2Id                  int    `json:"spell2Id"`
		Stats                     struct {
			Assists                         int  `json:"assists"`
			CausedEarlySurrender            bool `json:"causedEarlySurrender"`
			ChampLevel                      int  `json:"champLevel"`
			CombatPlayerScore               int  `json:"combatPlayerScore"`
			DamageDealtToObjectives         int  `json:"damageDealtToObjectives"`
			DamageDealtToTurrets            int  `json:"damageDealtToTurrets"`
			DamageSelfMitigated             int  `json:"damageSelfMitigated"`
			Deaths                          int  `json:"deaths"`
			DoubleKills                     int  `json:"doubleKills"`
			EarlySurrenderAccomplice        bool `json:"earlySurrenderAccomplice"`
			FirstBloodAssist                bool `json:"firstBloodAssist"`
			FirstBloodKill                  bool `json:"firstBloodKill"`
			FirstInhibitorAssist            bool `json:"firstInhibitorAssist"`
			FirstInhibitorKill              bool `json:"firstInhibitorKill"`
			FirstTowerAssist                bool `json:"firstTowerAssist"`
			FirstTowerKill                  bool `json:"firstTowerKill"`
			GameEndedInEarlySurrender       bool `json:"gameEndedInEarlySurrender"`
			GameEndedInSurrender            bool `json:"gameEndedInSurrender"`
			GoldEarned                      int  `json:"goldEarned"`
			GoldSpent                       int  `json:"goldSpent"`
			InhibitorKills                  int  `json:"inhibitorKills"`
			Item0                           int  `json:"item0"`
			Item1                           int  `json:"item1"`
			Item2                           int  `json:"item2"`
			Item3                           int  `json:"item3"`
			Item4                           int  `json:"item4"`
			Item5                           int  `json:"item5"`
			Item6                           int  `json:"item6"`
			KillingSprees                   int  `json:"killingSprees"`
			Kills                           int  `json:"kills"`
			LargestCriticalStrike           int  `json:"largestCriticalStrike"`
			LargestKillingSpree             int  `json:"largestKillingSpree"`
			LargestMultiKill                int  `json:"largestMultiKill"`
			LongestTimeSpentLiving          int  `json:"longestTimeSpentLiving"`
			MagicDamageDealt                int  `json:"magicDamageDealt"`
			MagicDamageDealtToChampions     int  `json:"magicDamageDealtToChampions"`
			MagicalDamageTaken              int  `json:"magicalDamageTaken"`
			NeutralMinionsKilled            int  `json:"neutralMinionsKilled"`
			NeutralMinionsKilledEnemyJungle int  `json:"neutralMinionsKilledEnemyJungle"`
			NeutralMinionsKilledTeamJungle  int  `json:"neutralMinionsKilledTeamJungle"`
			ObjectivePlayerScore            int  `json:"objectivePlayerScore"`
			ParticipantId                   int  `json:"participantId"`
			PentaKills                      int  `json:"pentaKills"`
			Perk0                           int  `json:"perk0"`
			Perk0Var1                       int  `json:"perk0Var1"`
			Perk0Var2                       int  `json:"perk0Var2"`
			Perk0Var3                       int  `json:"perk0Var3"`
			Perk1                           int  `json:"perk1"`
			Perk1Var1                       int  `json:"perk1Var1"`
			Perk1Var2                       int  `json:"perk1Var2"`
			Perk1Var3                       int  `json:"perk1Var3"`
			Perk2                           int  `json:"perk2"`
			Perk2Var1                       int  `json:"perk2Var1"`
			Perk2Var2                       int  `json:"perk2Var2"`
			Perk2Var3                       int  `json:"perk2Var3"`
			Perk3                           int  `json:"perk3"`
			Perk3Var1                       int  `json:"perk3Var1"`
			Perk3Var2                       int  `json:"perk3Var2"`
			Perk3Var3                       int  `json:"perk3Var3"`
			Perk4                           int  `json:"perk4"`
			Perk4Var1                       int  `json:"perk4Var1"`
			Perk4Var2                       int  `json:"perk4Var2"`
			Perk4Var3                       int  `json:"perk4Var3"`
			Perk5                           int  `json:"perk5"`
			Perk5Var1                       int  `json:"perk5Var1"`
			Perk5Var2                       int  `json:"perk5Var2"`
			Perk5Var3                       int  `json:"perk5Var3"`
			PerkPrimaryStyle                int  `json:"perkPrimaryStyle"`
			PerkSubStyle                    int  `json:"perkSubStyle"`
			PhysicalDamageDealt             int  `json:"physicalDamageDealt"`
			PhysicalDamageDealtToChampions  int  `json:"physicalDamageDealtToChampions"`
			PhysicalDamageTaken             int  `json:"physicalDamageTaken"`
			PlayerAugment1                  int  `json:"playerAugment1"`
			PlayerAugment2                  int  `json:"playerAugment2"`
			PlayerAugment3                  int  `json:"playerAugment3"`
			PlayerAugment4                  int  `json:"playerAugment4"`
			PlayerAugment5                  int  `json:"playerAugment5"`
			PlayerAugment6                  int  `json:"playerAugment6"`
			PlayerScore0                    int  `json:"playerScore0"`
			PlayerScore1                    int  `json:"playerScore1"`
			PlayerScore2                    int  `json:"playerScore2"`
			PlayerScore3                    int  `json:"playerScore3"`
			PlayerScore4                    int  `json:"playerScore4"`
			PlayerScore5                    int  `json:"playerScore5"`
			PlayerScore6                    int  `json:"playerScore6"`
			PlayerScore7                    int  `json:"playerScore7"`
			PlayerScore8                    int  `json:"playerScore8"`
			PlayerScore9                    int  `json:"playerScore9"`
			PlayerSubteamId                 int  `json:"playerSubteamId"`
			QuadraKills                     int  `json:"quadraKills"`
			SightWardsBoughtInGame          int  `json:"sightWardsBoughtInGame"`
			SubteamPlacement                int  `json:"subteamPlacement"`
			TeamEarlySurrendered            bool `json:"teamEarlySurrendered"`
			TimeCCingOthers                 int  `json:"timeCCingOthers"`
			TotalDamageDealt                int  `json:"totalDamageDealt"`
			TotalDamageDealtToChampions     int  `json:"totalDamageDealtToChampions"`
			TotalDamageTaken                int  `json:"totalDamageTaken"`
			TotalHeal                       int  `json:"totalHeal"`
			TotalMinionsKilled              int  `json:"totalMinionsKilled"`
			TotalPlayerScore                int  `json:"totalPlayerScore"`
			TotalScoreRank                  int  `json:"totalScoreRank"`
			TotalTimeCrowdControlDealt      int  `json:"totalTimeCrowdControlDealt"`
			TotalUnitsHealed                int  `json:"totalUnitsHealed"`
			TripleKills                     int  `json:"tripleKills"`
			TrueDamageDealt                 int  `json:"trueDamageDealt"`
			TrueDamageDealtToChampions      int  `json:"trueDamageDealtToChampions"`
			TrueDamageTaken                 int  `json:"trueDamageTaken"`
			TurretKills                     int  `json:"turretKills"`
			UnrealKills                     int  `json:"unrealKills"`
			VisionScore                     int  `json:"visionScore"`
			VisionWardsBoughtInGame         int  `json:"visionWardsBoughtInGame"`
			WardsKilled                     int  `json:"wardsKilled"`
			WardsPlaced                     int  `json:"wardsPlaced"`
			Win                             bool `json:"win"`
		} `json:"stats"`
		TeamId   models.TeamID `json:"teamId"`
		Timeline struct {
			CreepsPerMinDeltas struct {
			} `json:"creepsPerMinDeltas"`
			CsDiffPerMinDeltas struct {
			} `json:"csDiffPerMinDeltas"`
			DamageTakenDiffPerMinDeltas struct {
			} `json:"damageTakenDiffPerMinDeltas"`
			DamageTakenPerMinDeltas struct {
			} `json:"damageTakenPerMinDeltas"`
			GoldPerMinDeltas struct {
			} `json:"goldPerMinDeltas"`
			Lane               string `json:"lane"`
			ParticipantId      int    `json:"participantId"`
			Role               string `json:"role"`
			XpDiffPerMinDeltas struct {
			} `json:"xpDiffPerMinDeltas"`
			XpPerMinDeltas struct {
			} `json:"xpPerMinDeltas"`
		} `json:"timeline"`
	} `json:"participants"`
	PlatformId string `json:"platformId"`
	QueueId    int    `json:"queueId"`
	SeasonId   int    `json:"seasonId"`
	Teams      []struct {
		Bans                 []interface{} `json:"bans"`
		BaronKills           int           `json:"baronKills"`
		DominionVictoryScore int           `json:"dominionVictoryScore"`
		DragonKills          int           `json:"dragonKills"`
		FirstBaron           bool          `json:"firstBaron"`
		FirstBlood           bool          `json:"firstBlood"`
		FirstDargon          bool          `json:"firstDargon"`
		FirstInhibitor       bool          `json:"firstInhibitor"`
		FirstTower           bool          `json:"firstTower"`
		HordeKills           int           `json:"hordeKills"`
		InhibitorKills       int           `json:"inhibitorKills"`
		RiftHeraldKills      int           `json:"riftHeraldKills"`
		TeamId               int           `json:"teamId"`
		TowerKills           int           `json:"towerKills"`
		VilemawKills         int           `json:"vilemawKills"`
		Win                  string        `json:"win"`
	} `json:"teams"`
}

type FriendInfo struct {
	AccountId             int64               `json:"accountId"`
	Availability          models.FriendStatus `json:"availability"`
	GameName              string              `json:"gameName"`
	GameTag               string              `json:"gameTag"`
	Icon                  int                 `json:"icon"`
	Id                    string              `json:"id"`
	LegendaryMasteryScore int                 `json:"legendaryMasteryScore"`
	Lol                   struct {
		ChampionId            string `json:"championId"`
		CompanionId           string `json:"companionId"`
		DamageSkinId          string `json:"damageSkinId"`
		GameId                string `json:"gameId"`
		GameMode              string `json:"gameMode"`
		GameQueueType         string `json:"gameQueueType"`
		GameStatus            string `json:"gameStatus"`
		IconOverride          string `json:"iconOverride"`
		InitRankStat          string `json:"initRankStat"`
		IsObservable          string `json:"isObservable"`
		LegendaryMasteryScore string `json:"legendaryMasteryScore"`
		Level                 string `json:"level"`
		MapId                 string `json:"mapId"`
		MapSkinId             string `json:"mapSkinId"`
		ProfileIcon           string `json:"profileIcon"`
		Pty                   string `json:"pty"`
		Puuid                 string `json:"puuid"`
		QueueId               string `json:"queueId"`
		Regalia               string `json:"regalia"`
		SkinVariant           string `json:"skinVariant"`
		Skinname              string `json:"skinname"`
		TimeStamp             string `json:"timeStamp"`
	} `json:"lol"`
	MasteryScore             int           `json:"masteryScore"`
	Name                     string        `json:"name"`
	Note                     string        `json:"note"`
	PartySummoners           []interface{} `json:"partySummoners"`
	Patchline                string        `json:"patchline"`
	PlatformId               string        `json:"platformId"`
	Product                  string        `json:"product"`
	ProductName              string        `json:"productName"`
	Puuid                    string        `json:"puuid"`
	RemotePlatform           bool          `json:"remotePlatform"`
	RemoteProduct            bool          `json:"remoteProduct"`
	RemoteProductBackdropUrl string        `json:"remoteProductBackdropUrl"`
	RemoteProductIconUrl     string        `json:"remoteProductIconUrl"`
	StatusMessage            string        `json:"statusMessage"`
	SummonerIcon             int           `json:"summonerIcon"`
	SummonerId               int64         `json:"summonerId"`
	SummonerLevel            int           `json:"summonerLevel"`
}

type UserId struct {
	SummonerId int64  `json:"summonerId"`
	Puuid      string `json:"puuid"`
}

type UserName struct {
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type TeamInfo struct {
	UserList []UserId      `json:"userList"`
	TeamId   models.TeamID `json:"teamId"`
}

type ChampionSkinInfo struct {
	ChampionId int64 `json:"championId"`
	SkinId     int64 `json:"skinId"`
}

type SkinUrl struct {
	LoadScreenPath string `json:"loadScreenPath"`
}

type GameHistory struct {
	CreateTime int64  `json:"createTime"`
	GameId     int64  `json:"gameId"`
	GameMode   string `json:"gameMode"`
	GameType   string `json:"gameType"`
	ChampionId int64  `json:"championId"`
	QueueId    int64  `json:"queueId"`
	Win        bool   `json:"win"`
	Assists    int    `json:"assists"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
}

const (
	JoinedRoomMsg                                  = "joined_room"
	ConversationMsgTypeSystem ConversationMsgType  = "system"
	ChampSelectPatchTypePick  ChampSelectPatchType = "pick"
	ChampSelectPatchTypeBan   ChampSelectPatchType = "ban"
)

func (g GameInfo) ToGameHistory() GameHistory {
	if len(g.Participants) == 0 {
		syslog.L.Infof("没有参与比赛的数据", zap.Int64("gameId", g.GameId))
		return GameHistory{}
	}
	participant := g.Participants[0]
	return GameHistory{
		CreateTime: g.GameCreation,
		GameId:     g.GameId,
		GameMode:   string(g.GameMode),
		GameType:   string(g.GameType),
		ChampionId: int64(participant.ChampionId),
		Win:        participant.Stats.Win,
		Assists:    participant.Stats.Assists,
		Kills:      participant.Stats.Kills,
		Deaths:     participant.Stats.Deaths,
		QueueId:    int64(g.QueueId),
	}
}

type Lobby struct {
	CanStartActivity bool `json:"canStartActivity"`
	GameConfig       struct {
		AllowablePremadeSizes              []int         `json:"allowablePremadeSizes"`
		CustomLobbyName                    string        `json:"customLobbyName"`
		CustomMutatorName                  string        `json:"customMutatorName"`
		CustomRewardsDisabledReasons       []interface{} `json:"customRewardsDisabledReasons"`
		CustomSpectatorPolicy              string        `json:"customSpectatorPolicy"`
		CustomSpectators                   []interface{} `json:"customSpectators"`
		CustomTeam100                      []interface{} `json:"customTeam100"`
		CustomTeam200                      []interface{} `json:"customTeam200"`
		GameMode                           string        `json:"gameMode"`
		IsCustom                           bool          `json:"isCustom"`
		IsLobbyFull                        bool          `json:"isLobbyFull"`
		IsTeamBuilderManaged               bool          `json:"isTeamBuilderManaged"`
		MapId                              int           `json:"mapId"`
		MaxHumanPlayers                    int           `json:"maxHumanPlayers"`
		MaxLobbySize                       int           `json:"maxLobbySize"`
		MaxTeamSize                        int           `json:"maxTeamSize"`
		PickType                           string        `json:"pickType"`
		PremadeSizeAllowed                 bool          `json:"premadeSizeAllowed"`
		QueueId                            int           `json:"queueId"`
		ShouldForceScarcePositionSelection bool          `json:"shouldForceScarcePositionSelection"`
		ShowPositionSelector               bool          `json:"showPositionSelector"`
		ShowQuickPlaySlotSelection         bool          `json:"showQuickPlaySlotSelection"`
	} `json:"gameConfig"`
	Invitations []struct {
		InvitationId   string `json:"invitationId"`
		InvitationType string `json:"invitationType"`
		State          string `json:"state"`
		Timestamp      string `json:"timestamp"`
		ToSummonerId   int64  `json:"toSummonerId"`
		ToSummonerName string `json:"toSummonerName"`
	} `json:"invitations"`
	LocalMember struct {
		AllowedChangeActivity         bool        `json:"allowedChangeActivity"`
		AllowedInviteOthers           bool        `json:"allowedInviteOthers"`
		AllowedKickOthers             bool        `json:"allowedKickOthers"`
		AllowedStartActivity          bool        `json:"allowedStartActivity"`
		AllowedToggleInvite           bool        `json:"allowedToggleInvite"`
		AutoFillEligible              bool        `json:"autoFillEligible"`
		AutoFillProtectedForPromos    bool        `json:"autoFillProtectedForPromos"`
		AutoFillProtectedForRemedy    bool        `json:"autoFillProtectedForRemedy"`
		AutoFillProtectedForSoloing   bool        `json:"autoFillProtectedForSoloing"`
		AutoFillProtectedForStreaking bool        `json:"autoFillProtectedForStreaking"`
		BotChampionId                 int         `json:"botChampionId"`
		BotDifficulty                 string      `json:"botDifficulty"`
		BotId                         string      `json:"botId"`
		BotPosition                   string      `json:"botPosition"`
		BotUuid                       string      `json:"botUuid"`
		FirstPositionPreference       string      `json:"firstPositionPreference"`
		IntraSubteamPosition          interface{} `json:"intraSubteamPosition"`
		IsBot                         bool        `json:"isBot"`
		IsLeader                      bool        `json:"isLeader"`
		IsSpectator                   bool        `json:"isSpectator"`
		PlayerSlots                   []struct {
			ChampionId         int    `json:"championId"`
			Perks              string `json:"perks"`
			PositionPreference string `json:"positionPreference"`
			SkinId             int    `json:"skinId"`
			Spell1             int    `json:"spell1"`
			Spell2             int    `json:"spell2"`
		} `json:"playerSlots"`
		Puuid                    string      `json:"puuid"`
		QuickplayPlayerState     interface{} `json:"quickplayPlayerState"`
		Ready                    bool        `json:"ready"`
		SecondPositionPreference string      `json:"secondPositionPreference"`
		ShowGhostedBanner        bool        `json:"showGhostedBanner"`
		StrawberryMapId          interface{} `json:"strawberryMapId"`
		SubteamIndex             interface{} `json:"subteamIndex"`
		SummonerIconId           int         `json:"summonerIconId"`
		SummonerId               int64       `json:"summonerId"`
		SummonerInternalName     string      `json:"summonerInternalName"`
		SummonerLevel            int         `json:"summonerLevel"`
		SummonerName             string      `json:"summonerName"`
		TeamId                   int         `json:"teamId"`
		TftNPEQueueBypass        bool        `json:"tftNPEQueueBypass"`
	} `json:"localMember"`
	Members []struct {
		AllowedChangeActivity         bool        `json:"allowedChangeActivity"`
		AllowedInviteOthers           bool        `json:"allowedInviteOthers"`
		AllowedKickOthers             bool        `json:"allowedKickOthers"`
		AllowedStartActivity          bool        `json:"allowedStartActivity"`
		AllowedToggleInvite           bool        `json:"allowedToggleInvite"`
		AutoFillEligible              bool        `json:"autoFillEligible"`
		AutoFillProtectedForPromos    bool        `json:"autoFillProtectedForPromos"`
		AutoFillProtectedForRemedy    bool        `json:"autoFillProtectedForRemedy"`
		AutoFillProtectedForSoloing   bool        `json:"autoFillProtectedForSoloing"`
		AutoFillProtectedForStreaking bool        `json:"autoFillProtectedForStreaking"`
		BotChampionId                 int         `json:"botChampionId"`
		BotDifficulty                 string      `json:"botDifficulty"`
		BotId                         string      `json:"botId"`
		BotPosition                   string      `json:"botPosition"`
		BotUuid                       string      `json:"botUuid"`
		FirstPositionPreference       string      `json:"firstPositionPreference"`
		IntraSubteamPosition          interface{} `json:"intraSubteamPosition"`
		IsBot                         bool        `json:"isBot"`
		IsLeader                      bool        `json:"isLeader"`
		IsSpectator                   bool        `json:"isSpectator"`
		PlayerSlots                   []struct {
			ChampionId         int    `json:"championId"`
			Perks              string `json:"perks"`
			PositionPreference string `json:"positionPreference"`
			SkinId             int    `json:"skinId"`
			Spell1             int    `json:"spell1"`
			Spell2             int    `json:"spell2"`
		} `json:"playerSlots"`
		Puuid                    string      `json:"puuid"`
		QuickplayPlayerState     interface{} `json:"quickplayPlayerState"`
		Ready                    bool        `json:"ready"`
		SecondPositionPreference string      `json:"secondPositionPreference"`
		ShowGhostedBanner        bool        `json:"showGhostedBanner"`
		StrawberryMapId          interface{} `json:"strawberryMapId"`
		SubteamIndex             interface{} `json:"subteamIndex"`
		SummonerIconId           int         `json:"summonerIconId"`
		SummonerId               int64       `json:"summonerId"`
		SummonerInternalName     string      `json:"summonerInternalName"`
		SummonerLevel            int         `json:"summonerLevel"`
		SummonerName             string      `json:"summonerName"`
		TeamId                   int         `json:"teamId"`
		TftNPEQueueBypass        bool        `json:"tftNPEQueueBypass"`
	} `json:"members"`
	MucJwtDto struct {
		ChannelClaim string `json:"channelClaim"`
		Domain       string `json:"domain"`
		Jwt          string `json:"jwt"`
		TargetRegion string `json:"targetRegion"`
	} `json:"mucJwtDto"`
	MultiUserChatId       string        `json:"multiUserChatId"`
	MultiUserChatPassword string        `json:"multiUserChatPassword"`
	PartyId               string        `json:"partyId"`
	PartyType             string        `json:"partyType"`
	Restrictions          []interface{} `json:"restrictions"`
	ScarcePositions       []interface{} `json:"scarcePositions"`
	Warnings              []interface{} `json:"warnings"`
}
type SkinInfo struct {
	Active             bool          `json:"active"`
	Alias              string        `json:"alias"`
	BanVoPath          string        `json:"banVoPath"`
	BaseLoadScreenPath string        `json:"baseLoadScreenPath"`
	BaseSplashPath     string        `json:"baseSplashPath"`
	BotEnabled         bool          `json:"botEnabled"`
	ChooseVoPath       string        `json:"chooseVoPath"`
	DisabledQueues     []interface{} `json:"disabledQueues"`
	FreeToPlay         bool          `json:"freeToPlay"`
	Id                 int           `json:"id"`
	Name               string        `json:"name"`
	Ownership          struct {
		LoyaltyReward bool `json:"loyaltyReward"`
		Owned         bool `json:"owned"`
		Rental        struct {
			EndDate           int   `json:"endDate"`
			PurchaseDate      int64 `json:"purchaseDate"`
			Rented            bool  `json:"rented"`
			WinCountRemaining int   `json:"winCountRemaining"`
		} `json:"rental"`
		XboxGPReward bool `json:"xboxGPReward"`
	} `json:"ownership"`
	Passive struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	} `json:"passive"`
	Purchased         int64    `json:"purchased"`
	RankedPlayEnabled bool     `json:"rankedPlayEnabled"`
	Roles             []string `json:"roles"`
	Skins             []struct {
		ChampionId int     `json:"championId"`
		ChromaPath *string `json:"chromaPath"`
		Chromas    []struct {
			ChampionId   int      `json:"championId"`
			ChromaPath   string   `json:"chromaPath"`
			Colors       []string `json:"colors"`
			Disabled     bool     `json:"disabled"`
			Id           int      `json:"id"`
			LastSelected bool     `json:"lastSelected"`
			Name         string   `json:"name"`
			Ownership    struct {
				LoyaltyReward bool `json:"loyaltyReward"`
				Owned         bool `json:"owned"`
				Rental        struct {
					EndDate           int  `json:"endDate"`
					PurchaseDate      int  `json:"purchaseDate"`
					Rented            bool `json:"rented"`
					WinCountRemaining int  `json:"winCountRemaining"`
				} `json:"rental"`
				XboxGPReward bool `json:"xboxGPReward"`
			} `json:"ownership"`
			SkinAugments struct {
				Augments []interface{} `json:"augments"`
			} `json:"skinAugments"`
			StillObtainable bool `json:"stillObtainable"`
		} `json:"chromas"`
		CollectionSplashVideoPath interface{} `json:"collectionSplashVideoPath"`
		Disabled                  bool        `json:"disabled"`
		Emblems                   []struct {
			EmblemPath struct {
				Large string `json:"large"`
				Small string `json:"small"`
			} `json:"emblemPath"`
			Name      string `json:"name"`
			Positions struct {
				Horizontal string `json:"horizontal"`
				Vertical   string `json:"vertical"`
			} `json:"positions"`
		} `json:"emblems"`
		FeaturesText   interface{} `json:"featuresText"`
		Id             int         `json:"id"`
		IsBase         bool        `json:"isBase"`
		LastSelected   bool        `json:"lastSelected"`
		LoadScreenPath string      `json:"loadScreenPath"`
		Name           string      `json:"name"`
		Ownership      struct {
			LoyaltyReward bool `json:"loyaltyReward"`
			Owned         bool `json:"owned"`
			Rental        struct {
				EndDate           int   `json:"endDate"`
				PurchaseDate      int64 `json:"purchaseDate"`
				Rented            bool  `json:"rented"`
				WinCountRemaining int   `json:"winCountRemaining"`
			} `json:"rental"`
			XboxGPReward bool `json:"xboxGPReward"`
		} `json:"ownership"`
		QuestSkinInfo struct {
			CollectionCardPath    string        `json:"collectionCardPath"`
			CollectionDescription string        `json:"collectionDescription"`
			DescriptionInfo       []interface{} `json:"descriptionInfo"`
			Name                  string        `json:"name"`
			ProductType           interface{}   `json:"productType"`
			SplashPath            string        `json:"splashPath"`
			Tiers                 []interface{} `json:"tiers"`
			TilePath              string        `json:"tilePath"`
			UncenteredSplashPath  string        `json:"uncenteredSplashPath"`
		} `json:"questSkinInfo"`
		RarityGemPath string `json:"rarityGemPath"`
		SkinAugments  struct {
			Augments []interface{} `json:"augments"`
		} `json:"skinAugments"`
		SkinType             string      `json:"skinType"`
		SplashPath           string      `json:"splashPath"`
		SplashVideoPath      interface{} `json:"splashVideoPath"`
		StillObtainable      bool        `json:"stillObtainable"`
		TilePath             string      `json:"tilePath"`
		UncenteredSplashPath string      `json:"uncenteredSplashPath"`
	} `json:"skins"`
	Spells []struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	} `json:"spells"`
	SquarePortraitPath string `json:"squarePortraitPath"`
	StingerSfxPath     string `json:"stingerSfxPath"`
	TacticalInfo       struct {
		DamageType string `json:"damageType"`
		Difficulty int    `json:"difficulty"`
		Style      int    `json:"style"`
	} `json:"tacticalInfo"`
	Title string `json:"title"`
}
