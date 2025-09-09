package queue

const (
	QueueCatalogProcess = "catalog-process"
)

type EsChargeMessage struct {
	UpdatesCompleted bool   `json:"updatesCompleted"`
	Text             string `json:"text"`
	Env              string `json:"env"`
}

type CatalogProcessMessage struct {
	Id        int    `json:"id"`
	MediaType string `json:"mediaType"`
}
