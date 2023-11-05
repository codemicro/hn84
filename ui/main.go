package main

import (
	"git.tdpain.net/codemicro/hn84/ui/internal/config"
	"git.tdpain.net/codemicro/hn84/ui/internal/database"
	"git.tdpain.net/codemicro/hn84/ui/internal/httpcore"
	"git.tdpain.net/codemicro/hn84/util"
	"log/slog"
)

func main() {
	if err := run(); err != nil {
		slog.Error("unhandled error", "error", err)
	}
}

func run() error {
	db, err := database.Setup(config.Get().DatabaseName)
	if err != nil {
		return util.Wrap("setup database", err)
	}

	return httpcore.ListenAndServe(db)
}
