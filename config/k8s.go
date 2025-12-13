//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(connectify-record-mysql:3308)/connectify",
	},
	Redis: RedisConfig{
		Addr: "connectify-record-redis:6379",
	},
}
