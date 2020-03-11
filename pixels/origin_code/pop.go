package main

import (
	"github.com/valyala/fasthttp"
	"math"
	"strings"
	"time"
)

func logPop(ctx *fasthttp.RequestCtx) {
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

	insertPop(userCountry, userCity, blockId, siteId, ip, ua, subId, trafficOwner)

	uniqueHash := GetMD5Hash("Fg_e3" + GetMD5Hash(ip) + GetMD5Hash(ua))
	uniqueKey := `message_day_usr_unique_pop_` + uniqueHash
	uniqueValue, _ := env.client.Get(uniqueKey).Result()
	if len(uniqueValue) == 0 {
		env.client.Set(uniqueKey, 1, time.Duration(math.Abs(time.Since(time.Now().AddDate(0, 0, +1)).Seconds()))*time.Second)
		insertUniquePop(userCountry, userCity, blockId, siteId, ip, ua, subId, trafficOwner)
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
	return
}
