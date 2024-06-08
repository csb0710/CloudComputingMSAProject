// SPDX-License-Identifier: Apache-2.0

package common

import (
	"os"
)

// Config structure
type Config struct {
	DBHost     string
	DBName     string
	DBPassword string
}

// Cfg is for global reference
var Cfg Config

// LoadEnvVars loads environment variables and stores them as global variable
func LoadEnvVars() (Config, error) {
	Cfg.DBHost = os.Getenv("DB_HOST")
	Cfg.DBName = os.Getenv("DB_NAME")
	Cfg.DBPassword = os.Getenv("DB_PASSWORD")

	return Cfg, nil
}
