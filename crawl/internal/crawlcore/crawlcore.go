package crawlcore

import (
	"context"
	"git.tdpain.net/codemicro/hn84/crawl/internal/database"
	"git.tdpain.net/codemicro/hn84/util"
	"github.com/bwmarrin/snowflake"
	"github.com/uptrace/bun"
	"net/url"
	"strings"
	"sync"
)

var snow, _ = snowflake.NewNode(1)

type CrawlCore struct {
	DB       *bun.DB
	workerWG sync.WaitGroup
}

func New(db *bun.DB) *CrawlCore {
	return &CrawlCore{
		DB: db,
	}
}

func (c *CrawlCore) AddSite(startURL string) error {
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		return util.Wrap("parse start URL", err)
	}

	site := &database.Site{
		ID:       snow.Generate(),
		Domain:   parsedURL.Host,
		StartURL: startURL,
	}

	if _, err := c.DB.NewInsert().Model(site).Exec(context.Background()); err != nil {
		return util.Wrap("insert new Site record", err)
	}
	return nil
}

func hasDomainBeenSeen(db bun.IDB, domain string) (bool, error) {
	domain = strings.ToLower(domain)
	n, err := db.NewSelect().Model((*database.SeenDomain)(nil)).Where("domain = ?", domain).Count(context.Background())
	if err != nil {
		return false, util.Wrap("select seen domains", err)
	}
	return n >= 1, nil
}

func markDomainSeen(db bun.IDB, domain string) error {
	domain = strings.ToLower(domain)
	if _, err := db.NewInsert().Model(&database.SeenDomain{
		Domain: domain,
	}).Ignore().Exec(context.Background()); err != nil {
		return util.Wrap("mark domain as seen", err)
	}
	return nil
}
