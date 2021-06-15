package main

import (
	"context"
	"github.com/go-kratos/kratos/pkg/cache/memcache"
	"github.com/go-kratos/kratos/pkg/cache/redis"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/database/sql"
	"github.com/go-kratos/kratos/pkg/sync/pipeline/fanout"
	"github.com/pkg/errors"
	"log"
)

// 需要抛出给上层，方便追踪，排查问题
// 官方解释：
// ErrNoRows is returned by Scan when QueryRow doesn't return a
// row. In such a case, QueryRow returns a placeholder *Row value that
// defers this error until a Scan.
// var ErrNoRows = errors.New("sql: no rows in result set")

// dao dao.
type dao struct {
	db         *sql.DB
	redis      *redis.Redis
	mc         *memcache.Memcache
	cache      *fanout.Fanout
	demoExpire int32
}

type Article struct {
	ID      int64
	Content string
	Author  string
}

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	//log.Println(cfg)
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

func (d *dao) RawArticle(ctx context.Context, id int64) (art *Article, err error) {
	// get data from db
	art = &Article{}
	row := d.db.QueryRow(ctx, "select id,content,author from article where id=?", id)
	err = row.Scan(&art.ID, &art.Content, &art.Author)
	if err == sql.ErrNoRows {
		err = errors.Wrap(err, "RawArticle Scan ErrNoRows")
	}
	return
}
func main() {
	log.SetFlags(log.Llongfile)
	paladin.DefaultClient, _ = paladin.NewFile("./configs")
	db, _, err := NewDB()
	if err != nil {
		log.Println(err)
		return
	}
	d := &dao{
		db: db,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	art, err := d.RawArticle(ctx, 1)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(art)
}
