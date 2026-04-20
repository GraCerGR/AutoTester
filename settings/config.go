package settings

import "time"

// Redis
const (
	RedisAddr       = "localhost:6379"
	RedisPassword   = "test1234"
	RedisDB         = 0
	RedisMaxRetries = 5
	RedisTimeout    = 5 * time.Second
	DialTimeout     = 30 * time.Second
)

// kafka
var KafkaBrokers = []string{"localhost:9092"}
const (
	KafkaAttemptsTopic     = "attempts-attempts"
	KafkaResultsTopic      = "attempts-results"
	KafkaPartitions        = 1
	KafkaReplicationFactor = 1
	KafkaGroup             = "attempts-router"
	KafkaPollInterval      = 1 * time.Second
)

// Папки для хранения загружаемых файлов и результатов
const (
	FolderSite     = "Results/Sites/Gits/"
	FolderSolution = "Results/Solutions/Gits/"
	FolderLog      = "Results/Logs/"
)

// Database
const PostgresLink = "postgres://postgres:postgres@localhost:5432/VKR?sslmode=disable"

// Selenium Grid
const (
	HubStatusURL             = "http://localhost:4444/status"
	HubURL                   = "http://localhost:4444"
	HubWaitTimeout           = 40 * time.Second
	HubWaitPollInterval      = 1 * time.Second
	SeleniumNodeChromeNumber = "5"
)

// Контейнеры
type Container struct {
	Name  string
	Stack string
}

var TestContainers = []Container{
	{"test-node1", "python"},
	{"test-node2", "python"},
	{"test-node3", "java"},
	{"test-node4", "java"},
	{"test-node5", "None"},
}

var SiteContainers = []Container{
	{"worker1", ""},
	{"worker2", ""},
	{"worker3", ""},
	{"worker4", ""},
	{"worker5", ""},
}
