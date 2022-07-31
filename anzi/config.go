package anzi

import (
	"runtime"

	"github.com/ducesoft/overlord/pkg/log"
	"github.com/ducesoft/overlord/proxy"
)

// Config is the struct which used by cmd/anzi
type Config struct {
	*log.Config
	Migrate *MigrateConfig `toml:"migrate"`
}

// MigrateConfig is the config file which nedd to read/write into target dir.
type MigrateConfig struct {
	From              []*proxy.ClusterConfig `toml:"from"`
	To                *proxy.ClusterConfig   `toml:"to"`
	MaxRDBConcurrency int                    `toml:"max_rdb_concurrency"`
}

// SetDefault migrate config
func (m *MigrateConfig) SetDefault() {
	if m.MaxRDBConcurrency == 0 {
		m.MaxRDBConcurrency = runtime.NumCPU()
	}
	for _, from := range m.From {
		from.SetDefault()
	}
	m.To.SetDefault()
}
