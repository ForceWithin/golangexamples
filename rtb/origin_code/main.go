package main

import (
	"database/sql"
	"github.com/go-ini/ini"
	"github.com/go-redis/redis"
	"github.com/oschwald/geoip2-golang"
	"github.com/valyala/fasthttp"
	"log"
	"os"
)

type Env struct {
	db     *sql.DB
	geoip  *geoip2.Reader
	client *redis.Client
}

var (
	env Env
)

// main function to boot up everything
func main() {
	//cfg, err := ini.Load("rtb.ini")
	cfg, err := ini.Load(IniPath)
	if err != nil {
		log.Print("Fail to read file:", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Section(cfg.Section("").Key("application").String()).Key("sql").String())

	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)

	if err != nil {
		log.Print(err.Error())
	}
	defer db.Close()
	dbGeo, err := geoip2.Open(cfg.Section(cfg.Section("").Key("application").String()).Key("geo").String())
	if err != nil {
		log.Println(err)
	}
	defer dbGeo.Close()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		//DialTimeout:  10 * time.Second,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
		PoolSize: 10,
		//PoolTimeout:  30 * time.Second,
	})
	defer client.Close()

	env = Env{db: db, geoip: dbGeo, client: client}

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/g/mgid":
			env.GetMarketGid(ctx)
		case "/g/grtb":
			env.GetGRtb(ctx)
		case "/g/rtb":
			env.GetGRtb(ctx)
		case "/g/openrtb":
			env.GetRtb(ctx)
		case "/g/kadam":
			env.GetKadam(ctx)
		case "/g/prop":
			env.GetProp(ctx)
		case "/g/user":
			env.GetUser(ctx)
		default:
			ctx.Error("", fasthttp.StatusBadRequest)
		}
	}
	log.Fatal(fasthttp.ListenAndServe(":8100", m))
}
