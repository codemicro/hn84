package main

import (
	"fmt"
	"git.tdpain.net/codemicro/hn84/ui/internal/config"
	"git.tdpain.net/codemicro/hn84/ui/internal/database"
	"git.tdpain.net/codemicro/hn84/ui/internal/search"
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

	query := search.PlaintextToTokens("reading list")

	matches, err := search.DoSearch(db, query)
	if err != nil {
		return util.Wrap("run search", err)
	}

	fmt.Println(query)

	for _, m := range matches {
		fmt.Println(m.Document.Title, m.Ranking)
	}

	return nil
}
