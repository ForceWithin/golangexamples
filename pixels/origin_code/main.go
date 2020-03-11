package main

import (
	"database/sql"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/oschwald/geoip2-golang"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"os"
	"time"
)

type Env struct {
	db      *sql.DB
	client  *redis.Client
	loc     *time.Location
	geoip   *geoip2.Reader
	logDate string
}

var (
	env   Env
	Info  *log.Logger
	Error *log.Logger
)

func initLog(
	infoHandle io.Writer,
	errorHandle io.Writer) {

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	cfg, err := ini.Load(IniPath)
	if err != nil {
		log.Print("Fail to read file:", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Section(cfg.Section("").Key("application").String()).Key("sql").String())
	if err != nil {
		log.Print(err)
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(5)

	//db.SetConnMaxLifetime(time.Minute*5);
	//db.SetMaxIdleConns(5);
	//db.SetMaxOpenConns(5);

	defer db.Close()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		//DialTimeout:  10 * time.Second,
		//ReadTimeout:  30 * time.Second,
		//WriteTimeout: 30 * time.Second,
		PoolSize: 10,
		//PoolTimeout:  30 * time.Second,
	})
	defer client.Close()
	loc, _ := time.LoadLocation("Europe/Moscow")
	fNameInfo := fmt.Sprintf(cfg.Section(cfg.Section("").Key("application").String()).Key("log_messege_send_path").String(), time.Now().In(loc).Format("2006-01-02"))
	fNameError := fmt.Sprintf(cfg.Section(cfg.Section("").Key("application").String()).Key("log_messege_error_path").String(), time.Now().In(loc).Format("2006-01-02"))

	if cfg.Section("").Key("application").String() == "prod" {
		fInfo, errOpen := os.OpenFile(fNameInfo, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if errOpen != nil {
			log.Print("error opening file: %v", errOpen)
		}
		defer fInfo.Close()

		fError, errOpen := os.OpenFile(fNameError, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if errOpen != nil {
			log.Print("error opening file: %v", errOpen)
		}
		defer fError.Close()
		initLog(fInfo, fError)
	} else {
		initLog(os.Stdout, os.Stdout)
	}

	env = Env{db: db, client: client, loc: loc}

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/gc/sub":
			logSubscribes(ctx)
		case "/gc/sb":
			logSubscribes(ctx)
		case "/gc/sh":
			logShows(ctx)
		case "/gc/rev":
			revertUser(ctx)
		case "/gc/deny":
			logDeny(ctx)
		case "/gc/pop":
			logPop(ctx)
		case "/gc/click":
			logClicks(ctx)
		default:
			ctx.Error("", fasthttp.StatusBadRequest)
		}
	}
	log.Fatal(fasthttp.ListenAndServe(":8300", m))

}
