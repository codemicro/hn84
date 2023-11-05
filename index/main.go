package main

import (
	"bytes"
	"context"
	"encoding/json"
	"git.tdpain.net/codemicro/hn84/index/internal/config"
	"git.tdpain.net/codemicro/hn84/index/internal/database"
	"git.tdpain.net/codemicro/hn84/util"
	"github.com/PuerkitoBio/goquery"
	"github.com/uptrace/bun"
	"github.com/zentures/porter2"
	"log/slog"
	"os"
	"path"
	"strings"
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

	if err := walkDir(db, config.Get().CrawlDataDir); err != nil {
		return err
	}

	return nil
}

func walkDir(db *bun.DB, dir string) error {
	de, err := os.ReadDir(dir)
	if err != nil {
		return util.Wrap("read data dir", err)
	}

	for _, entry := range de {
		name := entry.Name()
		if !strings.HasSuffix(name, "html") {
			continue
		}

		id := name[:len(name)-5]

		// Process tokens

		htmlContent, err := os.ReadFile(path.Join(dir, name))
		if err != nil {
			return util.Wrap("read HTML file", err)
		}

		plaintext, pageTitle, err := convertHTMLToPlaintext(string(htmlContent))
		if err != nil {
			return util.Wrap("convert HTML to plaintext", err)
		}

		plaintext = filterPlaintextCharacters(plaintext)
		tokens := tokenise(plaintext)
		tokens = filterStopwords(tokens)
		stemTokens(tokens)

		dbTokens := convertToDatabaseTokens(tokens, id)
		if _, err := db.NewInsert().Model(&dbTokens).Exec(context.Background()); err != nil {
			return util.Wrap("unable to insert tokens to database", err)
		}

		// Dump plaintext to file
		if err := os.WriteFile(path.Join(dir, id+".txt"), []byte(plaintext), 0466); err != nil {
			return util.Wrap("write plaintext", err)
		}

		// Read extra data
		var dat = struct {
			URL string
		}{}

		jsonBytes, err := os.ReadFile(path.Join(dir, id+".json"))
		if err != nil {
			return util.Wrap("read document info", err)
		}

		if err := json.Unmarshal(jsonBytes, &dat); err != nil {
			return util.Wrap("unmarshal document info", err)
		}

		if _, err := db.NewInsert().Model(&database.Document{
			ID:    id,
			URL:   dat.URL,
			Title: pageTitle,
		}).Exec(context.Background()); err != nil {
			return util.Wrap("insert document to database", err)
		}

		break
	}

	return nil
}

type intermediateToken struct {
	Val        string
	Start, End int
}

func convertHTMLToPlaintext(htmlStr string) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(htmlStr))
	if err != nil {
		return "", "", util.Wrap("load HTML into goquery", err)
	}

	var titleStr string
	title := doc.Find("title")
	if len(title.Nodes) != 0 {
		titleStr = strings.TrimSpace(title.Text())
	}

	return titleStr + " " + strings.TrimSpace(doc.Find("body").Text()), titleStr, nil
}

func tokenise(plaintext string) []*intermediateToken {
	previousSpace := -1
	var tok []*intermediateToken
	pln := len(plaintext)
	for i, char := range plaintext {
		if char == ' ' || i == pln-1 {
			end := i - 1
			if char != ' ' {
				end += 1
				i += 1
			}
			tok = append(tok, &intermediateToken{
				Val:   strings.ToLower(plaintext[previousSpace+1 : i]),
				Start: previousSpace + 1,
				End:   end,
			})
			previousSpace = i
		}
	}
	return tok
}

func filterStopwords(tokens []*intermediateToken) []*intermediateToken {
	n := 0
	for _, tok := range tokens {
		_, found := stopwords[tok.Val]
		if !found {
			tokens[n] = tok
			n += 1
		}
	}
	return tokens[:n]
}

func stemTokens(tokens []*intermediateToken) {
	for _, tok := range tokens {
		tok.Val = porter2.Stem(tok.Val)
	}
}

func convertToDatabaseTokens(tokens []*intermediateToken, documentID string) []*database.Token {
	var res []*database.Token
	for _, tok := range tokens {
		res = append(res, &database.Token{
			Token:      tok.Val,
			DocumentID: documentID,
			Start:      tok.Start,
			End:        tok.End,
		})
	}
	return res
}

func filterPlaintextCharacters(plaintext string) string {
	arr := []rune(plaintext)
	n := 0
	for _, char := range arr {
		if ('A' <= char && char <= 'Z') || ('a' <= char && char <= 'z') || char == ' ' || ('0' <= char && char <= '9') {
			arr[n] = char
			n += 1
		}
	}

	return strings.Join(strings.Fields(string(arr[:n])), " ")
}

var stopwords = map[string]struct{}{
	"the":  {},
	"be":   {},
	"to":   {},
	"of":   {},
	"and":  {},
	"a":    {},
	"in":   {},
	"that": {},
	"have": {},
	"I":    {},
	"it":   {},
	"for":  {},
	"not":  {},
	"on":   {},
	"with": {},
	"he":   {},
	"as":   {},
	"you":  {},
	"do":   {},
	"at":   {},
	"this": {},
	"but":  {},
	"his":  {},
	"by":   {},
	"from": {},
}
