package configs

type Provider interface {
	Env() string
	MongoURI() string
	MongoDatabase() string
	ElasticHost() string
	ElasticUser() string
	ElasticPassword() string
	RabbitMQ() RabbitMQConfig
}

type EnvProvider struct{}

func NewProvider() Provider {
	return EnvProvider{}
}

func (EnvProvider) Env() string {
	return GetEnv()
}

func (EnvProvider) MongoURI() string {
	return MongoURI()
}

func (EnvProvider) MongoDatabase() string {
	return MongoDatabase()
}

func (EnvProvider) ElasticHost() string {
	return GetElkHost()
}

func (EnvProvider) ElasticUser() string {
	return GetElkUser()
}

func (EnvProvider) ElasticPassword() string {
	return GetELKPassword()
}

func (EnvProvider) RabbitMQ() RabbitMQConfig {
	return GetRabbitMQEnv()
}
