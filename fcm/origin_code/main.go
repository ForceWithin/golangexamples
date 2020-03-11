package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/valyala/fasthttp"
	"html"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"time"
)

type Task struct {
	closed chan struct{}
	wg     sync.WaitGroup
	ticker *time.Ticker
}

type Env struct {
	db         *sql.DB
	client     *redis.Client
	loc        *time.Location
	clientHttp *fasthttp.Client
}

var (
	env         Env
	Info        *log.Logger
	Dev         *log.Logger
	Error       *log.Logger
	user_prefix = map[int64]string{
		1: "_1",
		2: "_2",
	}
	fcm_keys = map[int64]string{
		1: "1",
		2: "2",
	}
	tblPrefix = map[int64]string{
		1: "",
		2: "_2",
	}
	replaceTblSuffix = map[int64]string{
		1: "_1",
		2: "",
	}
	//iniPath = "/usr/home/projects/pushmessages/fcm/origin_code/fcm.ini"
	iniPath = "/usr/home/projects/pushapi/fcm/fcm.ini"
)

func initLog(
	infoHandle io.Writer,
	errorHandle io.Writer) {

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Dev = log.New(infoHandle,
		"DEV: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// main function to boot up everything
func main() {

	cfg, err := ini.Load(iniPath)
	if err != nil {
		log.Print("Fail to read file:", err)
		os.Exit(1)
	}

	db, err := sql.Open("mysql", cfg.Section(cfg.Section("").Key("application").String()).Key("sql").String())
	if err != nil {
		log.Print(err)
	}
	db.SetMaxIdleConns(1)
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

	env = Env{db: db, client: client, loc: loc, clientHttp: &fasthttp.Client{}}

	task := &Task{
		closed: make(chan struct{}),
		ticker: time.NewTicker(time.Second * 1),
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	for fcm_key, fcm_value := range fcm_keys {
		task.wg.Add(1)
		go func(fcm_key int64, fcm_value string) {
			defer task.wg.Done()
			task.Run(fcm_key, fcm_value, cfg.Section("").Key("application").String())
		}(fcm_key, fcm_value)
	}

	select {
	case sig := <-c:
		fmt.Printf("Got %s signal. Aborting...\n", sig)
		task.Stop()
	}
}

func (t *Task) Run(fcm_id int64, fcm_value string, enviroment string) {

	for {
		select {
		case <-t.closed:
			return
		case <-t.ticker.C:
			startWork(fcm_id, fcm_value, enviroment)
		}
	}
}
func (t *Task) Stop() {
	close(t.closed)
	t.wg.Wait()
}

func startWork(fcm_id int64, fcm_value string, enviroment string) {
	unix := time.Now().In(env.loc).Unix()
	workUIds, errUIds := getUserIds(unix, tblPrefix[fcm_id])
	if errUIds != nil {
		log.Print(errUIds)
	}
	if len(workUIds) > 0 {
		letsDance(workUIds, unix, fcm_id, fcm_value)
	}
	if enviroment == "dev" {
		log.Print(len(workUIds))
	}
	workUIds = nil
}

func letsDance(workUIds []string, unix int64, fcm_id int64, fcm_value string) {
	cfg, errCfg := ini.Load(iniPath)
	if errCfg != nil {
		os.Exit(1)
	}
	resoursesDomain := cfg.Section(cfg.Section("").Key("application").String()).Key("resourses_domain").String()
	domainSsl := cfg.Section(cfg.Section("").Key("application").String()).Key("domain_ssl").String()
	fcmUrl := cfg.Section("").Key("fcm_url").String()

	err := deleteShow(unix, tblPrefix[fcm_id])
	unixString := fmt.Sprintf("%6d", unix)
	if err != nil {
		log.Print(err)
	}
	if cfg.Section("").Key("application").String() != "dev" {
		fNameInfo := fmt.Sprintf(cfg.Section(cfg.Section("").Key("application").String()).Key("log_messege_send_path").String(), time.Now().In(env.loc).Format("2006-01-02"))
		fNameError := fmt.Sprintf(cfg.Section(cfg.Section("").Key("application").String()).Key("log_messege_error_path").String(), time.Now().In(env.loc).Format("2006-01-02"))

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

	messages, err := getMessages()

	if err != nil {
		Error.Print(err.Error())
	}
	if len(messages) == 0 {
		return
	}

	userIds := []string{}

	keys := make([]int, 0, len(messages))
	for k := range messages {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		if len(workUIds) == len(userIds) {
			messages = nil
			userIds = nil
			return
		}

		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(messages[k]), func(i, j int) { messages[k][i], messages[k][j] = messages[k][j], messages[k][i] })

		for _, message := range messages[k] {
			if len(workUIds) == len(userIds) {
				messages = nil
				userIds = nil
				return
			}

			where := "u.id NOT IN(SELECT user_id FROM advert_show WHERE advert_id = " + message.Id + ")"
			//where += fmt.Sprintf(" AND u.timestamp = %6d", unix)
			where += " AND u.date_send < " + unixString + ""
			where += " AND u.date_reg < " + unixString + ""
			where += " AND mb.message_string LIKE '%," + message.Id + ",%'"
			if len(workUIds) > 0 {
				where += " AND u.id IN (" + strings.Join(workUIds, ",") + ") "
			}
			if len(userIds) > 0 {
				where += " AND u.id NOT IN (" + strings.Join(userIds, ",") + ") "
			}

			if len(message.Traffic_browsers) > 0 {
				where += fmt.Sprintf(" AND u.traffic_browsers IN(%s)", message.Traffic_browsers)
			}
			if len(message.Traffic_devices) > 0 {
				where += fmt.Sprintf(" AND u.traffic_devices IN(%s)", message.Traffic_devices)
			}

			where += " AND u.country_code IN (" + message.Country_string + ") "
			if message.City_string != "''" && len(message.City_string) > 0 {
				where += " AND IF(u.country_code IN('" + strings.Join(arCountriesSupportingCities, "','") + "'), u.city_code IN(" + message.City_string + "), u.country_code IN (" + message.Country_string + ")) "
			}

			//where += " AND IF(u.country_code IN('" + strings.Join(arCountriesSupportingCities, "','") + "'), " +
			//	"u.country_code IN(" + message.Country_string + ")  AND u.city_code IN(" + message.City_string + "), " +
			//	"u.country_code IN(" + message.Country_string + "))"

			users, err := getUsers(where, tblPrefix[fcm_id])

			if err != nil {
				Error.Print(err.Error())

			}

			if len(users) == 0 {
				continue
			}

			icon := message.Icon_url
			image := message.Image_url
			if message.Content_type == "local" {
				icon = resoursesDomain + message.Id + `/icon/` + message.Icon
				image = resoursesDomain + message.Id + `/image/` + message.Image
			}

			for _, userObjs := range users {
				ids := []string{}
				usersPerBlock := []User{}

				firstUser := userObjs[0]
				if message.Is_internal == 0 {
					cpcDb, _ := GetCpc(`id_block IN ('`+firstUser.Id_block+`', '-1') AND pay_type = '`+firstUser.Pay_type+`' `, `ORDER BY FIELD(id_block, `+firstUser.Id_block+`,-1)`)
					for _, userObj := range userObjs {

						contryCPC := ``
						cpc := 0.00
						if len(userObj.Traffic_devices) > 0 {
							contryCPC = GetMobileGeoGroupCountry(userObj.Country_code)
						} else {
							contryCPC = GetPcGeoGroupCountry(userObj.Country_code)
						}
						cpcBlock, okBlock := cpcDb[userObj.Id_block][contryCPC]
						if !okBlock {
							cpcDefault, okDefault := cpcDb[`-1`][contryCPC]
							if okDefault {
								cpc = cpcDefault.Cpc
							}
						} else {
							cpc = cpcBlock.Cpc
						}

						if cpc > 0 {
							if cpc < message.Cpc*1.2 {
								ids = append(ids, userObj.Id_user)
								usersPerBlock = append(usersPerBlock, userObj)
							}
						}

					}
				} else {
					for _, userObj := range userObjs {
						ids = append(ids, userObj.Id_user)
						usersPerBlock = append(usersPerBlock, userObj)
					}
				}

				for i := 0; i < len(ids); i += 1000 {
					batchVAlues := usersPerBlock[i:min(i+1000, len(usersPerBlock))]
					clickAction := domainSsl + `/c/c.php?u=` +
						base64.StdEncoding.EncodeToString([]byte(message.Url)) +
						`&s=` + firstUser.Id_site +
						`&b=` + firstUser.Id_block +
						`&msg=` + message.Id +
						`&fc` +
						`&rnd=` + unixString
					if message.Is_internal == 0 {
						click := Click{
							Url:        message.Url,
							BlockId:    firstUser.Id_block,
							MessageId:  message.Id,
							IsInternal: message.Is_internal,
							Random:     unixString,
							Cpc:        message.Cpc,
						}

						jsonclick, _ := json.Marshal(click)

						clickAction = domainSsl + `/c/clf.php?h=` +
							OpenSSLEncrypt(string(jsonclick))
					}

					request_body := RequestBody{
						RegistrationIds: ids[i:min(i+1000, len(ids))],
						Data: Fcm{
							Title:              html.UnescapeString(message.Title),
							Dir:                "ltr",
							RequireInteraction: true,
							Body:               html.UnescapeString(message.Body),
							Icon:               icon,
							Image:              image,
							Data: Data{
								Click_action: clickAction,
								Show_action: domainSsl + `/c/sh.php?s=` + firstUser.Id_site + `&b=` + firstUser.Id_block +
									`&msg=` + message.Id +
									`&sp=` + message.Show_period +
									`&rnd=` + unixString,
							},
						},
						Priority:     "high",
						CollapseKey:  "",
						Vibrate:      [3]int{100, 50, 100},
						Time_to_live: 60,
					}

					if len(message.Traffic_devices) > 0 {
						request_body.CollapseKey = "test"
					}

					//log.Println(string(js))
					ResponseData, errCall := callToFcm(request_body, fcm_value, fcmUrl)
					if errCall != nil {
						Error.Print(errCall.Error())
						batchVAlues = nil
						continue
					}

					for key, ResultFcm := range ResponseData.Results {

						ResultFcm.UserId = batchVAlues[key].User_id
						ResultFcm.FcmId = batchVAlues[key].FcmId

						if ResultFcm.MessageId == nil && ResultFcm.Error == nil {
							continue
						}
						userIdOriginal := GetMD5Hash(batchVAlues[key].Id_user_md5 + user_prefix[batchVAlues[key].FcmId])
						if ResultFcm.MessageId != nil {
							ResultFcm.Send = time.Now().In(env.loc).Format("2006-01-02 15:04:05")
							ResultFcm.Success = true
							Info.Printf("%+v\n", ResultFcm)
							err := insertSend(batchVAlues[key], message.Id, userIdOriginal)

							if err != nil {
								Error.Print(err.Error())
							}
							timerSend, _ := cfg.Section("").Key("send_interval").Int64()
							if len(batchVAlues[key].Traffic_browsers) > 0 {
								timerSend, _ = cfg.Section("").Key("send_interval_hour").Int64()
							}
							added := time.Now().In(env.loc).Add(time.Duration(timerSend) * time.Minute)

							errUpdate := updateUsersSend(added.Unix(), batchVAlues[key].User_id, tblPrefix[fcm_id])

							if errUpdate != nil {
								Error.Print(errUpdate.Error())
							}

							seen, err := env.client.Get("uniq_send_" + batchVAlues[key].User_id).Result()

							if len(seen) == 0 {
								env.client.Set("uniq_send_"+batchVAlues[key].User_id, 1, time.Duration(math.Abs(time.Since(time.Now().AddDate(0, 0, +1)).Seconds()))*time.Second)
								err := insertUniqueSend(batchVAlues[key], message.Id, userIdOriginal)
								if err != nil {
									Error.Print(err.Error())
								}
							}
						} else if ResultFcm.Error != nil && strings.Contains(strings.ToLower(*ResultFcm.Error), "register") {
							ResultFcm.Send = time.Now().In(env.loc).Format("2006-01-02 15:04:05")
							ResultFcm.Success = false
							Info.Printf("%+v\n", ResultFcm)
							err := InsertDeletedUserById(batchVAlues[key].User_id, tblPrefix[fcm_id])
							if err != nil {
								Error.Print(err.Error())
							}

							errD := deleteUserById(batchVAlues[key].User_id, tblPrefix[fcm_id])
							if errD != nil {
								Error.Print(errD.Error())
							}
							errLd := insertDeletes(batchVAlues[key], userIdOriginal)

							if errLd != nil {
								Error.Print(errLd.Error())
							}

						} else if ResultFcm.Error != nil && strings.Contains(strings.ToLower(*ResultFcm.Error), "mismatchsenderid") {
							ResultFcm.Send = time.Now().In(env.loc).Format("2006-01-02 15:04:05")
							ResultFcm.Success = false
							Info.Printf("%+v\n", ResultFcm)
							err := replaceUser(batchVAlues[key].User_id, replaceTblSuffix[fcm_id], tblPrefix[fcm_id])
							if err != nil {
								Error.Print(err.Error())
							}

							errD := deleteUserById(batchVAlues[key].User_id, tblPrefix[fcm_id])
							if errD != nil {
								Error.Print(errD.Error())
							}
						} else if cfg.Section("").Key("application").String() != "prod" && ResultFcm.Error != nil {
							ResultFcm.Send = time.Now().In(env.loc).Format("2006-01-02 15:04:05")
							ResultFcm.Success = false
							Info.Printf("%+v\n", ResultFcm)
							err := InsertDeletedUserById(batchVAlues[key].User_id, tblPrefix[fcm_id])
							if err != nil {
								Error.Print(err.Error())
							}

							errD := deleteUserById(batchVAlues[key].User_id, tblPrefix[fcm_id])
							if errD != nil {
								Error.Print(errD.Error())
							}
							errLd := insertDeletes(batchVAlues[key], userIdOriginal)

							if errLd != nil {
								Error.Print(errLd.Error())
							}
						}

					}
					ResponseData = ResultsFcm{}
					batchVAlues = nil
				}
				usersPerBlock = nil
				ids = nil
			}

		}

	}
	cfg = nil
	messages = nil
	userIds = nil
	return
}
