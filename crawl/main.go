package main

import (
	"errors"
	"git.tdpain.net/codemicro/hn84/crawl/internal/config"
	"git.tdpain.net/codemicro/hn84/crawl/internal/crawlcore"
	"git.tdpain.net/codemicro/hn84/crawl/internal/database"
	"git.tdpain.net/codemicro/hn84/util"
	"log/slog"
	"os"
)

func main() {
	if err := run(); err != nil {
		slog.Error("unrecoverable runtime error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	db, err := database.Setup(config.Get().DatabaseName)
	if err != nil {
		return util.Wrap("setup database", err)
	}

	cc := crawlcore.New(db)

	if len(os.Args) < 2 {
		return errors.New("too few arguments")
	}

	op := os.Args[1]

	switch op {
	case "add":
		if len(os.Args) < 3 {
			return errors.New("missing start URL")
		}

		if err := cc.AddSite(os.Args[2]); err != nil {
			return util.Wrap("add site", err)
		}
	case "run":
		if err := cc.Loop(nil); err != nil {
			return util.Wrap("run crawl loop", err)
		}
	default:
		return errors.New("unrecognised operation")
	}

	return nil
}
