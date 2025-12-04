package settings

import "time"

// Redis
var RedisAddr = "localhost:6379"
var RedisPassword = "test1234"
var RedisDB = 0
var RedisMaxRetries = 5
var RedisTimeout = 5 * time.Second
var DialTimeout = 30 * time.Second

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
