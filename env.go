package gracehttp

import (
	"os"
)

const (
	env_SERVER_RESTART_PREFIX = "gracehttp_server_restart_"
)

func setRestartEnv(key string) {
	os.Setenv(env_SERVER_RESTART_PREFIX+key, "1")
}

func isRestartEnv(key string) bool {
	envVal := os.Getenv(env_SERVER_RESTART_PREFIX + key)
	return envVal == "1"
}
