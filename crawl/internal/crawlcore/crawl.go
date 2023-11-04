package crawlcore

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"git.tdpain.net/codemicro/hn84/crawl/internal/config"
	"git.tdpain.net/codemicro/hn84/crawl/internal/database"
	"git.tdpain.net/codemicro/hn84/util"
	"github.com/PuerkitoBio/goquery"
	"github.com/carlmjohnson/requests"
	"log/slog"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"
	"time"
)

func (c *CrawlCore) Loop(stop chan struct{}) error {
	jobs := make(chan *database.Site)

	defer func() {
		slog.Info("waiting for all workers to terminate")
		c.workerWG.Wait()
	}()
	defer close(jobs)

	numWorkers := config.Get().NumWorkers

	_ = os.MkdirAll(config.Get().CrawlDataDir, os.ModeDir)

	slog.Info("starting workers", "n", numWorkers)

	for i := 0; i < numWorkers; i += 1 {
		go c.worker(i, jobs)
	}

	slog.Info("crawl loop alive")

mainLoop:
	for {
		if stop != nil {
			select {
			case <-stop:
				break mainLoop
			default:
			}
		}

		tx, err := c.DB.BeginTx(context.Background(), nil)
		if err != nil {
			return util.Wrap("create crawl loop transaction", err)
		}

		site := new(database.Site)
		if err := tx.NewSelect().Model(site).Order("id asc").Limit(1).Scan(context.Background(), site); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tx.Rollback()
				time.Sleep(time.Second)
				continue
			}
			tx.Rollback()
			return util.Wrap("select next site to crawl", err)
		}

		if _, err := tx.NewDelete().Model(site).Where("id = ?", site.ID).Exec(context.Background()); err != nil {
			tx.Rollback()
			return util.Wrap("delete crawled site", err)
		}

		if err := markDomainSeen(tx, site.Domain); err != nil {
			tx.Rollback()
			return util.Wrap("mark domain as seen", err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return util.Wrap("commit crawl loop transaction", err)
		}

		jobs <- site
	}

	return nil
}

type pageMeta struct {
	Time time.Time
	URL  string
}

func (c *CrawlCore) worker(workerID int, jobChan chan *database.Site) {
	c.workerWG.Add(1)

	log := slog.With("worker", workerID)

	for site := range jobChan {
		conf := config.Get()
		log.Info("run", "domain", site.Domain)
		currPageNumber := 0

		if site.StartURL == "" {
			site.StartURL = "http://" + site.Domain + "/"
		}

		queuedURLs := map[string]struct{}{
			site.StartURL: {},
		}
		urlQueue := []string{site.StartURL}

		otherDomains := make(map[string]string)

	pageLoop:
		for {
			if currPageNumber == conf.MaxPagesPerDomain {
				log.Info("hit page number limit", "domain", site.Domain)
				break
			}
			var currentURL string
			if len(urlQueue) == 0 {
				break
			} else {
				currentURL = urlQueue[0]
				urlQueue = slices.Delete(urlQueue, 0, 1)
			}

			// Get page
			var pageBody string

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			err := requests.URL(currentURL).ToString(&pageBody).UserAgent(conf.UserAgent).Fetch(ctx)
			cancel()

			if err != nil {
				log.Warn("failed to fetch page", "url", currentURL, "error", err)
				break pageLoop
			}

			// Extract links
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(pageBody)))
			if err != nil {
				// TODO: Check that we get HTML back
				log.Error("failed to parse page", "error", err, "url", currentURL)
				continue pageLoop
			}

			currentParsedURL, err := url.Parse(currentURL)
			if err != nil {
				log.Error("unable to parse current page URL", "error", err, "url", currentURL)
				continue pageLoop
			}

			doc.Find("a").Each(func(i int, selection *goquery.Selection) {
				href, exists := selection.Attr("href")
				if !exists || href == "" {
					return
				}

				thisURL, err := url.Parse(href)
				if err != nil {
					return
				}

				if thisURL.Scheme == "" {
					thisURL.Scheme = currentParsedURL.Scheme
				} else if !(thisURL.Scheme == "http" || thisURL.Scheme == "https") {
					return
				}

				if thisURL.Host == "" {
					if thisURL.Path == "" {
						// This is either a URL with a fragment or query string
						return
					}
					thisURL.Host = currentParsedURL.Host

					if !strings.HasPrefix(thisURL.Path, "/") {
						if strings.HasSuffix(currentParsedURL.Path, "/") {
							thisURL.Path = currentParsedURL.Path + thisURL.Path
						} else {
							thisURL.Path = path.Base(currentParsedURL.Path) + "/" + thisURL.Path
						}
					}
				} else {
					// TODO: this is potentially fumbling some pages that aren't linked directly within the other website

					tx, err := c.DB.Begin()
					if err != nil {
						log.Error("faild to open transaction", "error", err)
						return
					}
					defer tx.Rollback()

					host := strings.ToLower(thisURL.Host)
					if _, found := otherDomains[host]; found {
						return
					}

					if beenSeen, err := hasDomainBeenSeen(tx, host); err != nil {
						log.Error("failed to lookup domain status", "error", err)
						return
					} else if beenSeen {
						return
					}

					otherDomains[host] = href

					if err := markDomainSeen(tx, host); err != nil {
						log.Error("failed to mark domain as seen", "error", err)
						return
					}

					if _, err := tx.NewInsert().Model(&database.Site{
						ID:     snow.Generate(),
						Domain: host,
					}).Exec(context.Background()); err != nil {
						log.Error("insert new Site record", "error", err)
						return
					}

					if err := tx.Commit(); err != nil {
						log.Error("failed to commit cross-domain link checking transaction", "error", err)
						return
					}

					return
				}

				thisURL.RawQuery = ""
				thisURL.Fragment = ""

				u := thisURL.String()
				if _, found := queuedURLs[u]; !found {
					urlQueue = append(urlQueue, u)
					queuedURLs[u] = struct{}{}
				}
			})

			// Dump to disk
			metaData := &pageMeta{
				Time: time.Now(),
				URL:  currentURL,
			}
			metaJSON, err := json.Marshal(metaData)
			if err != nil {
				log.Error("failed to marshal site metadata", "error", err)
				break
			}

			baseFname := path.Join(conf.CrawlDataDir, fmt.Sprintf("%d.%d", site.ID, currPageNumber))

			if err := os.WriteFile(baseFname+".html", []byte(pageBody), 0644); err != nil {
				log.Error("failed to dump page content", "error", err, "fname", baseFname)
			}

			if err := os.WriteFile(baseFname+".json", metaJSON, 0644); err != nil {
				log.Error("failed to dump page metadata", "error", err, "fname", baseFname)
			}

			// Go places
			currPageNumber += 1
		}
	}

	log.Info("worker stopping")

	c.workerWG.Done()
}
