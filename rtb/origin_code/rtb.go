package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-ini/ini"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"log"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type SeatBidRtb struct {
	Bids []BidRTb `json:"bid"`
}

type answerRtb struct {
	Cur     string       `json:"cur"`
	BidId   int          `json:"bidid"`
	Seatbid []SeatBidRtb `json:"seatbid"`
}

type rtbPost struct {
	Device Device `json:"device"`
	Site   Site   `json:"site"`
}

type BidRTb struct {
	Id    string  `json:"id"`
	Nurl  string  `json:"nurl"`
	Price float64 `json:"price"`
	Adm   AdmType `json:"adm"`
}

type AdmType struct {
	Icon    string `json:"icon"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Url     string `json:"url"`
}

func (env *Env) GetRtb(ctx *fasthttp.RequestCtx) {
	defer ctx.Response.ConnectionClose()
	var rtbPostData rtbPost

	errDecode := json.Unmarshal(ctx.PostBody(), &rtbPostData)
	if errDecode != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	key := string(ctx.QueryArgs().Peek("key"))

	if len(key) == 0 {
		//log.Println("Url Param 'key' is missing")
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	ua := rtbPostData.Device.Ua

	if len(ua) < 1 {
		//log.Println("Url Param 'ua' is missing")
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	userIp := rtbPostData.Device.Ip

	if len(userIp) < 1 {
		//log.Println("Url Param 'user_ip' is missing")
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	postCountry := rtbPostData.Device.Geo.Country
	userCountry := "UN"
	if len(postCountry) < 1 || postCountry == "{country_iso_2}" {
		// If you are using strings that may be invalid, check that ip is not nil
		ip := net.ParseIP(userIp)
		country, err := env.geoip.City(ip)
		if err != nil {
			log.Println(err)
		}
		userCountry = string(country.Country.IsoCode)
	} else {
		userCountry = postCountry
	}

	stream := rtbPostData.Site.Id

	if len(stream) < 1 || stream == "{src_id}" {
		stream = "0"
	}

	if env.BlockBlackListSubidByBlockAndSubid(key, stream) {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	if env.BlockWhiteListSubidByBlockAndSubid(key, stream) {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	uniqueHash := GetMD5Hash("Fg_e3" + GetMD5Hash(userIp) + GetMD5Hash(ua))

	seen, _ := env.client.Get(uniqueHash + "unique_rtb" + key).Result()

	if len(seen) > 0 && userIp != "193.105.200.212" {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	errCC := env.GetCountryByCode(userCountry)

	if errCC != nil {
		userCountry = "UN"
	}

	isMobileDevice := IsMobile(ua)
	device := ``
	browser := ``
	if isMobileDevice {
		device = GetDevice(ua)
	} else {
		browser = GetBrowserByUa(ua)
	}

	block, errBlock := env.GetBlockById(key)

	if errBlock != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	cfg, err := ini.Load(IniPath)
	if err != nil {
		log.Print("Fail to read file:", err)
		os.Exit(1)
	}

	timeOut, _ := cfg.Section("").Key("timeout_rtb").Int()

	if block.UniqeShowsRtb > 0 {
		timeOut = block.UniqeShowsRtb * 60 * 60
	}
	env.client.Set(uniqueHash+"unique_rtb"+key, 1, time.Duration(timeOut)*time.Second)

	if len(block.MessageString) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	formatedNow := now.Format("2006-01-02 15:04:05")

	userMessageTimeShowsCookieName := `message_inf_` + key + `_` + uniqueHash

	userMessageShowInfo, _ := env.client.Get(userMessageTimeShowsCookieName).Result()

	maxExpireTimestamp := time.Now().AddDate(0, 0, +1)

	messageTime := map[string]string{}
	json.Unmarshal([]byte(userMessageShowInfo), &messageTime)

	messageTimeForQuery := []string{}
	if len(messageTime) > 0 {
		for index, element := range messageTime {
			if element <= formatedNow {
				delete(messageTime, index)
			} else {
				if element > maxExpireTimestamp.Format("2006-01-02 00:00:00") {
					maxExpireTimestamp, _ = time.Parse("2006-01-02 15:04:05", element)
				}
				messageTimeForQuery = append(messageTimeForQuery, index)
			}
		}
	}

	whereMessages := ``
	whereMessages += "id IN (" + strings.Trim(block.MessageString, ",") + ") "

	if len(messageTimeForQuery) > 0 {
		whereMessages += "AND id NOT IN (" + strings.Join(messageTimeForQuery, ",") + ") "
	}
	whereMessages += "AND country_string LIKE '%" + userCountry + "%' "
	if isMobileDevice {
		whereMessages += "AND traffic_devices LIKE '%" + device + "%' "
	} else {
		whereMessages += "AND traffic_devices = '' " // no mobile devices must be listed
		whereMessages += "AND (traffic_browsers LIKE '%" + browser + "%' OR traffic_browsers = '') "

	}
	message, errMessage := env.GetMessages(whereMessages)

	if errMessage != nil {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	}

	//messageTime
	if message.Show_period > 0 {
		now.Add(time.Duration(message.Show_period*60) * time.Minute)
		messageTime[message.Id] = now.Format("2006-01-02 15:04:05")
	}
	if len(messageTime) > 0 {
		sliceToRedis, _ := json.Marshal(messageTime)
		_ = env.client.Set(userMessageShowInfo, sliceToRedis, time.Duration(math.Abs(time.Since(maxExpireTimestamp).Seconds()))*time.Second)
	}
	shavingKey := `bl_cnt_` + key

	blockCounter, _ := env.client.Get(shavingKey).Result()
	//setPool2, erroroexp := poolConn.Do("EXPIRE", shavingKey, time.Now().AddDate(0, 0, +1).Unix())
	blockShows := 1
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

	contryCPC := ``
	if isMobileDevice {
		contryCPC = GetMobileGeoGroupCountry(userCountry)
	} else {
		contryCPC = GetPcGeoGroupCountry(userCountry)
	}

	cpc := 0.000
	cpcDb, errCpc := env.GetCpc(`id_block IN ('` + key + `', '-1') AND code = '` + contryCPC + `' AND pay_type = '` + block.PayType + `' `)
	if errCpc != nil || cpcDb.Cpc == 0 {
		ctx.SetStatusCode(fasthttp.StatusNoContent)
		return
	} else {
		cpc = cpcDb.Cpc
	}

	domainSsl := cfg.Section(cfg.Section("").Key("application").String()).Key("domain_ssl").String()
	resoursesDomain := cfg.Section(cfg.Section("").Key("application").String()).Key("resourses_domain").String()
	domainNoSsl := cfg.Section(cfg.Section("").Key("application").String()).Key("domain_no_ssl").String()

	icon := message.Icon_url
	//image := message.Image_url
	if message.Content_type == "local" {
		icon = resoursesDomain + message.Id + `/icon/` + message.Icon
		//image = resoursesDomain + message.Id + `/image/` + message.Image
	}

	//msgId, _ := strconv.Atoi(message.Id)
	randomHash := fmt.Sprintf("%6d", time.Now().Unix())
	intKey, _ := strconv.Atoi(key)
	answer := &answerRtb{
		Cur:   "RUB",
		BidId: intKey,
		Seatbid: []SeatBidRtb{
			{
				Bids: []BidRTb{
					{
						Id: "1",
						Nurl: domainNoSsl + Show + `?h=` + uniqueHash +
							`&s=` + block.IdSite + `&b=` + block.IdBlock +
							`&msg=` + message.Id + `&c=` + userCountry + `&cc=UN` +
							`&sv=` + strconv.Itoa(block.Shaving) +
							`&subid=` + stream + `&r=` + randomHash,
						Price: cpc,
						Adm: AdmType{
							Icon:    icon,
							Title:   message.Title,
							Content: message.Body,
							Url: domainSsl + Click + `?h=` + uniqueHash + `&u=` +
								base64.StdEncoding.EncodeToString([]byte(message.Url)) +
								`&s=` + block.IdSite + `&b=` + block.IdBlock +
								`&msg=` + message.Id + `&c=` + userCountry + `&cc=UN` +
								`&sv=` + strconv.Itoa(block.Shaving) +
								`&subid=` + stream + `&r=` + randomHash,
						},
					},
				},
			},
		},
	}

	js, _ := json.Marshal(answer)

	js = bytes.Replace(js, []byte("\\u003c"), []byte("<"), -1)
	js = bytes.Replace(js, []byte("\\u003e"), []byte(">"), -1)
	js = bytes.Replace(js, []byte("\\u0026"), []byte("&"), -1)

	//w.WriteHeader(200)
	ctx.Response.Header.Set("Content-Type", "application/json; charset=utf8")
	ctx.SetBody(js)

	if key == cfg.Section("").Key("debug_block").String() {
		fileName := fmt.Sprintf(cfg.Section(cfg.Section("").Key("application").String()).Key("log_path").String(), now.Format("2006-01-02")+"_"+key)
		f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		defer f.Close()
		log.SetOutput(f)

		postJson, _ := JSONMarshal(rtbPostData, true)
		log.Println("Request: key="+key, string(postJson))
		log.Println("Response: ", string(js))
	}

	//c.Data["json"] = answer
	//c.ServeJSON()
	return
}
