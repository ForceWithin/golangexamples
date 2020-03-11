package main

import (
	"encoding/base64"
	"github.com/valyala/fasthttp"
	"math"
	"strconv"
	"strings"
	"time"
)

func logClicks(ctx *fasthttp.RequestCtx) {
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

	url := string(params.Peek("u"))
	if len(url) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	UrlDecoded, errDecoding := base64.StdEncoding.DecodeString(url)
	if errDecoding != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}
	url = string(UrlDecoded)

	userId := string(params.Peek("uh"))
	if len(userId) == 0 {
		userId = ""
	} else {
		userId = GetMD5Hash(userId)
	}

	block, errBlock := GetBlockById(blockId)

	if errBlock != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	//cfg, _ := ini.Load(IniPath)

	blockShows := 1
	if block.PayType == "clicks" && block.Shaving < 100 {
		shavingKey := `bl_cnt_cl_` + blockId
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

	userIdOriginal := ""
	if len(userId) > 0 {
		userIdOriginal = GetMD5Hash(userId + user_prefix[block.FcmId])
	}

	for _, broker := range brokers {

		if strings.Contains(url, "%"+broker+"%") {
			domain, errDomain := GetDomaneByBrokerAndCountrey(broker, userCountry)
			if errDomain != nil {
				break
			}
			url = strings.ReplaceAll(url, "%"+broker+"%", domain)
			break
		}
	}
	deviceType := "pc"
	if IsMobile(ua) {
		deviceType = GetDevice(ua)
	}

	url = strings.ReplaceAll(url, "%subacc%", block.IdSite)
	url = strings.ReplaceAll(url, "%messageid%", messageId)
	url = strings.ReplaceAll(url, "%idad%", messageId)
	url = strings.ReplaceAll(url, "%idblock%", block.IdBlock)
	url = strings.ReplaceAll(url, "%geo%", userCountry)
	url = strings.ReplaceAll(url, "%ref%", string(ctx.Referer()))
	url = strings.ReplaceAll(url, "%subid%", subId)
	url = strings.ReplaceAll(url, "%type%", deviceToUrls[deviceType])
	url = strings.ReplaceAll(url, "%pushuserid%", userIdOriginal)

	if block.PayType == "clicks" && block.Shaving < 100 && blockShows > block.Shaving {
		insertClicksShaving(userCountry, userCity, block, ip, ua, subId, messageId, userIdOriginal)
	} else {
		insertClicks(userCountry, userCity, block, ip, ua, subId, messageId, userIdOriginal)
	}

	ctx.Redirect(url, fasthttp.StatusFound)
	return
}
