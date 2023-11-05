package config

import (
	"git.tdpain.net/pkg/cfger"
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	DatabaseName      string
	NumWorkers        int
	MaxPagesPerDomain int
	UserAgent         string
	CrawlDataDir      string
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
			DatabaseName:      cl.WithDefault("crawler.databaseName", "crawler.db").AsString(),
			NumWorkers:        cl.WithDefault("crawler.numWorkers", 8).AsInt(),
			MaxPagesPerDomain: cl.WithDefault("crawler.maxPagesPerDomain", 300).AsInt(),
			UserAgent:         cl.WithDefault("crawler.userAgent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0").AsString(),
			CrawlDataDir:      cl.WithDefault("dataDir", "crawlData").AsString(),
		}
	})

	if outerErr != nil {
		slog.Error("fatal error when loading configuration", "err", outerErr)
		os.Exit(1)
	}

	return conf
}
