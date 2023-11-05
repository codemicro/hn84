package search

import (
	"context"
	"git.tdpain.net/codemicro/hn84/ui/internal/database"
	"git.tdpain.net/codemicro/hn84/util"
	"github.com/uptrace/bun"
	"github.com/zentures/porter2"
	"math"
	"sort"
	"strings"
)

type Match struct {
	Document *database.Document
	Tokens   []*database.Token
	Ranking  float64
}

func DoSearch(db *bun.DB, query []string) ([]*Match, error) {
	query = util.Deduplicate(query)

	var tokens []*database.Token
	if err := db.NewSelect().Model(&tokens).Where("token in (?)", bun.In(query)).Scan(context.Background(), &tokens); err != nil {
		return nil, util.Wrap("unable to execute query on database", err)
	}

	tokensByDocument := make(map[string][]*database.Token)

	for _, token := range tokens {
		tokensByDocument[token.DocumentID] = append(tokensByDocument[token.DocumentID], token)
	}

	// each document must contain all tokens
	var docsToDelete []string
	for doc, tokens := range tokensByDocument {
		seen := make(map[string]struct{})
		for _, tok := range tokens {
			seen[tok.Token] = struct{}{}
		}
		if len(seen) != len(query) {
			docsToDelete = append(docsToDelete, doc)
		}
	}

	for _, doc := range docsToDelete {
		delete(tokensByDocument, doc)
	}

	termFrequencies := make(map[string]map[string]int)
	for doc, tokens := range tokensByDocument {
		freqs := make(map[string]int)
		for _, tok := range tokens {
			freqs[tok.Token] = freqs[tok.Token] + 1
		}
		termFrequencies[doc] = freqs
	}

	var totalNumDocs float64
	{
		tnd, err := db.NewSelect().Model((*database.Document)(nil)).Count(context.Background())
		if err != nil {
			return nil, util.Wrap("count all documents", err)
		}
		totalNumDocs = float64(tnd)
	}

	idfs := make(map[string]float64)
	for _, tokStr := range query {
		var occursInN int
		err := db.NewRaw(`SELECT COUNT(*) FROM (SELECT '' FROM tokens WHERE token = ? GROUP BY "document_id")`, tokStr).Scan(context.Background(), &occursInN)
		if err != nil {
			return nil, util.Wrap("count number of documents that term occurs in", err)
		}
		idfs[tokStr] = math.Log(totalNumDocs / float64(occursInN))
	}

	var res []*Match

	for docID, tokens := range tokensByDocument {
		doc := new(database.Document)
		if err := db.NewSelect().Model(doc).Where("id = ?", docID).Scan(context.Background(), doc); err != nil {
			return nil, util.Wrap("final assembly", err)
		}

		m := &Match{
			Document: doc,
			Tokens:   tokens,
		}

		freq := termFrequencies[docID]

		for _, tok := range tokens {
			m.Ranking += float64(freq[tok.Token]) * idfs[tok.Token]
		}

		res = append(res, m)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Ranking > res[j].Ranking
	})

	return res, nil
}

func PlaintextToTokens(plain string) []string {
	plain = filterPlaintextCharacters(plain)
	tokens := tokenise(plain)
	tokens = filterStopwords(tokens)
	stemTokens(tokens)
	return tokens
}

func tokenise(plaintext string) []string {
	previousSpace := -1
	var tok []string
	pln := len(plaintext)
	for i, char := range plaintext {
		if char == ' ' || i == pln-1 {
			end := i - 1
			if char != ' ' {
				end += 1
				i += 1
			}
			tok = append(tok, strings.ToLower(plaintext[previousSpace+1:i]))
			previousSpace = i
		}
	}
	return tok
}

func filterStopwords(tokens []string) []string {
	n := 0
	for _, tok := range tokens {
		_, found := stopwords[tok]
		if !found {
			tokens[n] = tok
			n += 1
		}
	}
	return tokens[:n]
}

func stemTokens(tokens []string) {
	for i, tok := range tokens {
		tokens[i] = porter2.Stem(tok)
	}
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
