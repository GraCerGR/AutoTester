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
var TestContainers = []string{"test-node1", "test-node2", "test-node3", "test-node4", "test-node5"}

var SiteContainers = []string{"site1", "site2", "site3", "site4", "site5"}
