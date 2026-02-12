package domain

type AutomationConfig struct {
	AutoConfirm bool `json:"autoConfirm"`
	AutoPick    int  `json:"autoPick"`
	AutoBan     int  `json:"autoBan"`
}
