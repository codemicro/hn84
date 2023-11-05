package database

import (
	"database/sql"
	"errors"
	"git.tdpain.net/codemicro/hn84/util"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"os"
)

type Document struct {
	bun.BaseModel

	ID    string `bun:",pk"`
	URL   string
	Title string
}

type Token struct {
	bun.BaseModel

	Token      string
	DocumentID string
	Start, End int
}

func Setup(filepath string) (*bun.DB, error) {
	alreadyExists := true
	if _, err := os.Stat(filepath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		alreadyExists = false
	}

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, util.Wrap("open database", err)
	}

	db.SetMaxOpenConns(1) // https://github.com/mattn/go-sqlite3/issues/274#issuecomment-191597862

	b := bun.NewDB(db, sqlitedialect.New())
	//b.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	if !alreadyExists {
		return nil, errors.New("cannot create database from new in ui")
	}

	return b, nil
}
