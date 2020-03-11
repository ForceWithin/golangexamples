package main

import (
	"github.com/go-ini/ini"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

func revertUser(ctx *fasthttp.RequestCtx) {
	//ctx.RemoteAddr()
	//ctx.Request.Header.VisitAll(func (key, value []byte) {
	//	log.Printf("%s: %s", string(key), string(value))
	//})

	defer ctx.Response.ConnectionClose()
	params := ctx.QueryArgs()

	siteId := string(params.Peek("sid"))
	if len(siteId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	blockId := string(params.Peek("bid"))
	if len(blockId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	userId := string(params.Peek("id"))
	if len(userId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	block, errBlock := GetBlockById(blockId)

	if errBlock != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	cfg, _ := ini.Load(IniPath)

	subId := string(params.Peek("subid"))

	if len(subId) == 0 {
		subId = "0"
	}

	trafficOwner := string(params.Peek("to"))

	if len(trafficOwner) == 0 || trafficOwner == "" {
		trafficOwner = "self"
	}

	userCountry := string(ctx.Request.Header.Peek("HTTP_GEOIP_COUNTRY_CODE"))

	if len(userCountry) == 0 || userCountry == "" {
		userCountry = "UN"
	}

	userCity := string(ctx.Request.Header.Peek("HTTP_GEOIP_CITY"))

	userCity = GetUserCity(userCountry, userCity)

	if _, ok := citiesLike[userCountry][strings.ToLower(userCity)]; ok {
		userCity = strings.ToUpper(citiesLike[userCountry][strings.ToLower(userCity)])
	}

	ip := string(ctx.Request.Header.Peek("X-Real-IP"))
	ua := string(ctx.Request.Header.Peek("User-Agent"))

	//uniqueHash := GetMD5Hash("Fg_e3" + GetMD5Hash(ip) + GetMD5Hash(ua))
	if len(ua) > 200 {
		runes := []rune(ua)

		ua = string(runes[0:200])
	}

	device := ""
	browser := ""

	if IsMobile(ua) {
		device = GetDevice(ua)
	} else {
		browser = GetBrowserByUa(ua)
	}

	deleteDeletedUsers(tbl_prefix[block.FcmId], GetMD5Hash(userId))

	timerSend, _ := cfg.Section("").Key("first_send_interval").Int64()

	timeout := time.Now().In(env.loc).Add(time.Duration(timerSend) * time.Minute)

	insertUser(tbl_prefix[block.FcmId], userId, timeout.UnixNano(), userCountry,
		userCity, block, ip, browser, device, ua, subId, trafficOwner)

	ctx.SetStatusCode(fasthttp.StatusNoContent)
	return
}
