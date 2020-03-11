package main

type Click struct {
	Url        string  `json:"u"`
	BlockId    string  `json:"b"`
	MessageId  string  `json:"msg"`
	IsInternal int     `json:"in"`
	Random     string  `json:"r"`
	Cpc        float64 `json:"cpc"`
}

type RequestBody struct {
	RegistrationIds []string `json:"registration_ids"`
	Data            Fcm      `json:"data"`
	Priority        string   `json:"priority"`
	CollapseKey     string   `json:"collapse_key"`
	Vibrate         [3]int   `json:"vibrate"`
	Time_to_live    int64    `json:"time_to_live"`
}

type Fcm struct {
	Title              string `json:"title"`
	Dir                string `json:"dir"`
	RequireInteraction bool   `json:"requireInteraction"`
	Body               string `json:"body"`
	Icon               string `json:"icon"`
	Image              string `json:"image"`
	Data               Data   `json:"data"`
}

type Data struct {
	Click_action string `json:"click_action"`
	Show_action  string `json:"show_action"`
}

type Message struct {
	Id               string
	Url              string
	Title            string
	Body             string
	Image            string
	Icon             string
	Image_url        string
	Icon_url         string
	Content_type     string
	Message_priority int
	Country_string   string
	City_string      string
	Traffic_browsers string
	Traffic_devices  string
	Show_period      string
	Is_internal      int
	Cpc              float64
}

type User = struct {
	User_id          string
	Id_user          string
	Id_user_md5      string
	Country_code     string
	City_code        string
	Id_block         string
	Ip               string
	User_agent       string
	Id_site          string
	Subid            string
	Traffic_owner    string
	Traffic_browsers string
	Traffic_devices  string
	Shaving          string
	Pay_type         string
	FcmId            int64
}

type ResultsFcm struct {
	Results []ResultFcm `json:"results"`
}

type ResultFcm struct {
	Error     *string `json:"error"`
	MessageId *string `json:"message_id"`
	UserId    string  `json:"user_id,omitempty"`
	Send      string  `json:"send,omitempty"`
	Success   bool    `json:"success,omitempty"`
	FcmId     int64   `json:"success,omitempty"`
}

type Cpc struct {
	Id_block     string
	Country_code string
	Cpc          float64
}

var (
	arCountriesSupportingCities = []string{
		"RU",
		"UA",
		"KZ",
	}
)

func deleteShow(time int64, tbl_prefix string) error {
	_, err := env.db.Exec("DELETE FROM advert_show"+tbl_prefix+" WHERE show_end < ?", time)

	if err != nil {
		return err
	}
	return nil
}

func getUserIds(time int64, tbl_prefix string) ([]string, error) {
	results, err := env.db.Query(
		"SELECT u.id FROM users"+tbl_prefix+" u "+
			"INNER JOIN message_blocks mb "+
			"ON u.id_block = mb.id_block "+
			"WHERE u.date_send < ? AND u.date_reg < ? order by u.date_send", time, time)

	var id string
	if err != nil {
		return nil, err
	}
	defer results.Close()
	users := []string{}
	for results.Next() {

		err = results.Scan(&id)

		if err != nil {
			return nil, err
		}

		users = append(users, id)
	}

	return users, nil

}

func getMessages() (map[int][]Message, error) {
	results, err := env.db.Query(
		"SELECT " +
			"id, url, title, body, image, message_priority," +
			"icon, image_url, icon_url, content_type, show_period, " +
			"country_string, city_string, traffic_browsers, traffic_devices, is_internal, cpc " +
			"FROM advert_message " +
			"ORDER BY message_priority")

	if err != nil {
		return nil, err
	}
	defer results.Close()
	message := Message{}
	messagesByPriority := map[int][]Message{}

	for results.Next() {

		err = results.Scan(&message.Id, &message.Url, &message.Title, &message.Body, &message.Image,
			&message.Message_priority, &message.Icon, &message.Image_url, &message.Icon_url,
			&message.Content_type, &message.Show_period, &message.Country_string, &message.City_string,
			&message.Traffic_browsers, &message.Traffic_devices, &message.Is_internal, &message.Cpc)

		if err != nil {
			return nil, err
		}

		messagesByPriority[message.Message_priority] = append(messagesByPriority[message.Message_priority], message)
	}

	message = Message{}

	return messagesByPriority, nil
}

func GetCpc(where string, orderby string) (map[string]map[string]Cpc, error) {

	results, err := env.db.Query(
		"SELECT id_block, code, cpc FROM wm_tariffs_message " +
			"WHERE " + where +
			" " + orderby)
	if err != nil {
		return nil, err
	}
	defer results.Close()
	cpcResult := map[string]map[string]Cpc{}

	cpc := Cpc{}
	for results.Next() {
		err = results.Scan(&cpc.Id_block, &cpc.Country_code, &cpc.Cpc)

		if err != nil {
			return nil, err
		}
		mm, ok := cpcResult[cpc.Id_block]
		if !ok {
			mm = make(map[string]Cpc)
			cpcResult[cpc.Id_block] = mm
		}
		mm[cpc.Country_code] = cpc
	}

	cpc = Cpc{}

	return cpcResult, nil
}

func getUsers(where string, tbl_prefix string) (map[string][]User, error) {

	results, err := env.db.Query(
		"SELECT " +
			"u.id user_id, u.id_user_md5, u.id_user, u.country_code, u.city_code, " +
			"u.id_block, u.ip, u.user_agent, u.id_site, " +
			"u.subid, u.traffic_owner, u.traffic_browsers, u.traffic_devices, mb.shaving, mb.fcm_id, mb.pay_type " +
			"FROM users" + tbl_prefix + " u " +
			"INNER JOIN message_blocks mb " +
			"ON u.id_block = mb.id_block " +
			"WHERE " + where + " order by u.date_send")

	if err != nil {
		return nil, err
	}
	defer results.Close()
	user := User{}
	users := map[string][]User{}
	for results.Next() {

		err = results.Scan(&user.User_id, &user.Id_user_md5, &user.Id_user, &user.Country_code, &user.City_code, &user.Id_block,
			&user.Ip, &user.User_agent, &user.Id_site, &user.Subid, &user.Traffic_owner, &user.Traffic_browsers, &user.Traffic_devices,
			&user.Shaving, &user.FcmId, &user.Pay_type)

		if err != nil {
			return nil, err
		}

		users[user.Id_block] = append(users[user.Id_block], user)
	}

	user = User{}

	return users, nil
}

func insertSend(user User, idMessage string, userId string) error {
	_, err := env.db.Exec("INSERT INTO log_sends ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"id_message, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"id_user "+
		") VALUES ("+
		"CURDATE(), "+
		"CURTIME(), ?, ?, ?,?,?,?,?,?,?)",
		user.Country_code,
		user.City_code,
		user.Id_site,
		user.Id_block,
		idMessage,
		user.Ip,
		user.User_agent,
		user.Subid,
		userId)

	return err
}

func updateUsersSend(sendTimeStamp int64, userId string, tbl_prefix string) error {
	_, err := env.db.Exec("UPDATE users"+tbl_prefix+" SET date_send = ? WHERE id = ?", sendTimeStamp, userId)

	return err
}

func insertUniqueSend(user User, idMessage string, userId string) error {
	_, err := env.db.Exec("INSERT INTO log_actives ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"id_message, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner, "+
		"id_user "+
		") VALUES ("+
		"CURDATE(), "+
		"CURTIME(), ?, ?, ?,?,?,?,?,?,?,?)",
		user.Country_code,
		user.City_code,
		user.Id_site,
		user.Id_block,
		idMessage,
		user.Ip,
		user.User_agent,
		user.Subid,
		user.Traffic_owner,
		userId)

	return err
}

func insertDeletes(user User, userId string) error {
	_, err := env.db.Exec("INSERT INTO log_deletes ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner, "+
		"id_user "+
		") VALUES ("+
		"CURDATE(), "+
		"CURTIME(), ?, ?, ?,?,?,?,?,?,?)",
		user.Country_code,
		user.City_code,
		user.Id_site,
		user.Id_block,
		user.Ip,
		user.User_agent,
		user.Subid,
		user.Traffic_owner,
		userId)

	return err
}

func deleteUserById(userId string, tbl_prefix string) error {
	_, err := env.db.Exec("DELETE FROM users"+tbl_prefix+" WHERE id = ?", userId)

	return err
}

func InsertDeletedUserById(userId string, tbl_prefix string) error {
	_, err := env.db.Exec("INSERT IGNORE INTO deleted_users"+tbl_prefix+" SELECT * FROM users"+tbl_prefix+" WHERE id = ?", userId)

	return err
}

func replaceUser(userId string, replace_tbl_suffix string, tbl_suffix string) error {
	_, err := env.db.Exec("INSERT IGNORE INTO users"+replace_tbl_suffix+" SELECT * FROM users"+tbl_suffix+" WHERE id = ?", userId)

	return err
}
