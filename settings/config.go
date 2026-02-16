package settings

import "time"

// Redis
const (
	RedisAddr = "localhost:6379"
	RedisPassword = "test1234"
	RedisDB = 0
	RedisMaxRetries = 5
	RedisTimeout = 5 * time.Second
	DialTimeout = 30 * time.Second
)

// Database
const PostgresLink = "postgres://postgres:postgres@localhost:5432/VKR?sslmode=disable"

// Selenium Grid
const (
	HubStatusURL        = "http://localhost:4444/status"
	HubURL              = "http://localhost:4444"
	HubWaitTimeout      = 40 * time.Second
	HubWaitPollInterval = 1 * time.Second
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
	{"site1", ""},
	{"site2", ""},
	{"site3", ""},
	{"site4", ""},
	{"site5", ""},
}
