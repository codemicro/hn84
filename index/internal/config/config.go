package config

import (
	"git.tdpain.net/pkg/cfger"
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	DatabaseName string
	CrawlDataDir string
}

var (
	conf     *Config
	loadOnce = new(sync.Once)
)

func Get() *Config {
	var outerErr error
	loadOnce.Do(func() {
		cl := cfger.New()
		if err := cl.Load("config.yml"); err != nil {
			outerErr = err
			return
		}

		conf = &Config{
			DatabaseName: cl.WithDefault("index.databaseName", "index.db").AsString(),
			CrawlDataDir: cl.WithDefault("dataDir", "crawlData").AsString(),
		}
	})

	if outerErr != nil {
		slog.Error("fatal error when loading configuration", "err", outerErr)
		os.Exit(1)
	}

	return conf
}
