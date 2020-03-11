package main

import (
	"github.com/valyala/fasthttp"
	"math"
	"strconv"
	"strings"
	"time"
)

func logShows(ctx *fasthttp.RequestCtx) {
	//ctx.RemoteAddr()
	//ctx.Request.Header.VisitAll(func (key, value []byte) {
	//	log.Printf("%s: %s", string(key), string(value))
	//})

	defer ctx.Response.ConnectionClose()
	params := ctx.QueryArgs()

	siteId := string(params.Peek("s"))
	if len(siteId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	blockId := string(params.Peek("b"))
	if len(blockId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	messageId := string(params.Peek("msg"))
	if len(messageId) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	userId := string(params.Peek("r"))
	if len(userId) == 0 {
		userId = ""
	} else {
		userId = GetMD5Hash(userId)
	}

	showPeriod, _ := strconv.Atoi(string(params.Peek("sp")))

	block, errBlock := GetBlockById(blockId)

	if errBlock != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	//cfg, _ := ini.Load(IniPath)

	blockShows := 1
	if block.PayType == "shows" && block.Shaving < 100 {
		shavingKey := `bl_cnt_` + blockId
		blockCounter, _ := env.client.Get(shavingKey).Result()
		if len(blockCounter) == 0 {
			env.client.Set(shavingKey, blockShows, time.Duration(math.Abs(time.Since(time.Now().AddDate(0, 0, +3)).Seconds()))*time.Second)
		} else {
			if blockCounter >= "100" {
				env.client.Set(shavingKey, blockShows, time.Duration(math.Abs(time.Since(time.Now().AddDate(0, 0, +3)).Seconds()))*time.Second)
			} else {
				env.client.Incr(shavingKey)
			}
			blockCounter, _ := strconv.Atoi(blockCounter)
			blockCounter++
			blockShows = blockCounter
		}
	}

	subId := string(params.Peek("subid"))

	if len(subId) == 0 {
		subId = "0"
	}

	userCountry := string(params.Peek("c"))
	if len(userCountry) == 0 {
		userCountry = string(ctx.Request.Header.Peek("HTTP_GEOIP_COUNTRY_CODE"))
	}
	if len(userCountry) == 0 || userCountry == "" {
		userCountry = "UN"
	}

	userCity := string(params.Peek("cc"))
	if len(userCity) == 0 {
		userCity = string(ctx.Request.Header.Peek("HTTP_GEOIP_CITY"))
	}
	userCity = GetUserCity(userCountry, userCity)

	if _, ok := citiesLike[userCountry][strings.ToLower(userCity)]; ok {
		userCity = strings.ToUpper(citiesLike[userCountry][strings.ToLower(userCity)])
	}
	ip := string(ctx.Request.Header.Peek("X-Real-IP"))
	ua := string(ctx.Request.Header.Peek("User-Agent"))
	uniqueHash := string(params.Peek("h"))
	if len(uniqueHash) == 0 {
		uniqueHash = GetMD5Hash("Fg_e3" + GetMD5Hash(ip) + GetMD5Hash(ua))
	}

	if len(ua) > 200 {
		runes := []rune(ua)

		ua = string(runes[0:200])
	}

	if showPeriod > 0 {
		expire := time.Now().In(env.loc).Add(time.Duration(showPeriod) * time.Hour)
		insertAdverShow(tbl_prefix[block.FcmId], messageId, expire.Unix(), userId)
	}

	userIdOriginal := ""
	if len(userId) > 0 {
		userIdOriginal = GetMD5Hash(userId + user_prefix[block.FcmId])
	}

	if block.PayType == "shows" && block.Shaving < 100 && blockShows > block.Shaving {
		insertShowsShaving(userCountry, userCity, block, ip, ua, subId, messageId, userIdOriginal)
	} else {
		insertShows(userCountry, userCity, block, ip, ua, subId, messageId, userIdOriginal)
	}

	uniqueKey := `message_day_usr_unique_shows_` + messageId + `_` + uniqueHash
	uniqueValue, _ := env.client.Get(uniqueKey).Result()
	if len(uniqueValue) == 0 {
		env.client.Set(uniqueKey, 1, time.Duration(math.Abs(time.Since(time.Now().AddDate(0, 0, +1)).Seconds()))*time.Second)
		insertUniqueShows(userCountry, userCity, block, ip, ua, subId, messageId, userIdOriginal)
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
	return
}
