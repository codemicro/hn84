package database

import (
	"context"
	"database/sql"
	"errors"
	"git.tdpain.net/codemicro/hn84/util"
	"github.com/bwmarrin/snowflake"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"os"
)

type Site struct {
	bun.BaseModel

	ID       snowflake.ID `bun:",pk"`
	Domain   string       `bun:",notnull,unique"`
	StartURL string
}

type SeenDomain struct {
	bun.BaseModel

	Domain string `bun:",pk"`
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
		if _, err := b.NewCreateTable().Model((*Site)(nil)).Exec(context.Background()); err != nil {
			return nil, util.Wrap("create Site table", err)
		}

		if _, err := b.NewCreateTable().Model((*SeenDomain)(nil)).Exec(context.Background()); err != nil {
			return nil, util.Wrap("create SeenDomain table", err)
		}
	}

	return b, nil
}
