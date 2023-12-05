package queue

type EsChargeMessage struct {
	UpdatesCompleted bool   `json:"updatesCompleted"`
	Text             string `json:"text"`
	Env              string `json:"env"`
}
