package main

import (
	"bytes"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"log"
	"net"
)

type UserDATA struct {
	Ua           string `json:"ua"`
	Geo          string `json:"geo"`
	IsMobile     bool   `json:"is_mobile"`
	MobileDevice string `json:"mobile_device"`
	PcDevice     string `json:"pc_device"`
}

func (env *Env) GetUser(ctx *fasthttp.RequestCtx) {
	defer ctx.Response.ConnectionClose()
	answer := &UserDATA{}
	ua := string(ctx.QueryArgs().Peek("ua"))
	answer.Ua = ua
	userIp := string(ctx.QueryArgs().Peek("ip"))
	geo := string(ctx.QueryArgs().Peek("c"))

	if len(geo) < 1 {
		// If you are using strings that may be invalid, check that ip is not nil
		ip := net.ParseIP(userIp)
		country, err := env.geoip.City(ip)
		if err != nil {
			log.Println(err)
		}
		geo = string(country.Country.IsoCode)
	}
	answer.Geo = geo
	isMobileDevice := IsMobile(ua)
	answer.IsMobile = isMobileDevice
	device := ``
	browser := ``
	if isMobileDevice {
		device = GetDevice(ua)
		answer.MobileDevice = device
	} else {
		browser = GetBrowserByUa(ua)
		answer.PcDevice = browser
	}

	js, _ := json.Marshal(answer)

	js = bytes.Replace(js, []byte("\\u003c"), []byte("<"), -1)
	js = bytes.Replace(js, []byte("\\u003e"), []byte(">"), -1)
	js = bytes.Replace(js, []byte("\\u0026"), []byte("&"), -1)

	//w.WriteHeader(200)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf8")
	ctx.SetBody(js)

	//c.Data["json"] = answer
	//c.ServeJSON()
	return
}
