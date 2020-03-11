package main

type Block struct {
	IdBlock       string
	IdSite        string
	PayType       string
	Shaving       int
	MessageString string
	UniqeShowsRtb int
	FcmId         int64
}

type User struct {
	IdUser int64
}

func GetBlockById(key string) (*Block, error) {
	block := &Block{}

	err := env.db.QueryRow(
		"SELECT id_block, id_site, pay_type, shaving, message_string, unique_shows_rtb, fcm_id "+
			"FROM message_blocks "+
			"WHERE id_block = ?", key).Scan(&block.IdBlock, &block.IdSite, &block.PayType,
		&block.Shaving, &block.MessageString, &block.UniqeShowsRtb, &block.FcmId)

	if err != nil {
		return nil, err
	}

	return block, nil
}

func GetDomaneByBrokerAndCountrey(broker string, country string) (string, error) {
	var domain string

	err := env.db.QueryRow(
		"SELECT domain "+
			"FROM projects_domains "+
			"WHERE broker = ? AND geo IN (?, 'ALL') ORDER BY FIELD (geo, ?, 'ALL') LIMIT 1", broker, country, country).Scan(&domain)

	if err != nil {
		return "", err
	}

	return domain, nil
}

func GetUserByMd5(tbl_prefix string, key string) (*User, error) {
	user := &User{}

	err := env.db.QueryRow(
		"SELECT id "+
			"FROM users"+tbl_prefix+" "+
			"WHERE id_user_md5 = ?", key).Scan(&user.IdUser)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func insertUser(prefix string, userId string, timeout int64, coutry string, city string, block *Block, ip string,
	browser string, device string, ua string, subid string, trafficOwner string) {
	_, err := env.db.Exec("INSERT IGNORE INTO users"+prefix+" ( "+
		"id_user, "+
		"id_user_md5, "+
		"date_reg, "+
		"date_send, "+
		"country_code, "+
		"city_code, "+
		"id_block, "+
		"id_site, "+
		"ip, "+
		"traffic_browsers, "+
		"traffic_devices, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner, "+
		"fcm_id "+
		") VALUES ("+
		"?, ?, ?,?,?,?,?,?,?,?,?,?,?,?,?)",
		userId, GetMD5Hash(userId), timeout, timeout, coutry, city, block.IdBlock, block.IdSite,
		ip, browser, device, ua, subid, trafficOwner, block.FcmId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return

}

func insertSubscriptionShaving(coutry string, city string, block *Block, ip string,
	ua string, subid string, userId string, trafficOwner string) {
	_, err := env.db.Exec("INSERT INTO log_subscribtions_shaving ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"id_user, "+
		"traffic_owner "+
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, ip, ua, subid, userId, trafficOwner)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertSubscription(coutry string, city string, block *Block, ip string,
	ua string, subid string, userId string, trafficOwner string) {
	_, err := env.db.Exec("INSERT INTO log_subscribtions ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"id_user, "+
		"traffic_owner "+
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, ip, ua, subid, userId, trafficOwner)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertAdverShow(prefix string, messageId string, expire int64, userId string) {
	_, err := env.db.Exec("INSERT IGNORE INTO advert_show"+prefix+
		" SELECT id, ? advert_id, ? show_end FROM users"+prefix+" WHERE id_user_md5 = ?",
		messageId, expire, userId)
	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertShowsShaving(coutry string, city string, block *Block, ip string, ua string, subid string, messageId string, userId string) {

	_, err := env.db.Exec("INSERT INTO log_shows_shaving ( "+
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
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, messageId, ip, ua, subid, userId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertShows(coutry string, city string, block *Block, ip string, ua string, subid string, messageId string, userId string) {
	_, err := env.db.Exec("INSERT INTO log_shows ( "+
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
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, messageId, ip, ua, subid, userId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertUniqueShows(coutry string, city string, block *Block, ip string, ua string, subid string, messageId string, userId string) {
	_, err := env.db.Exec("INSERT INTO log_shows_unique ( "+
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
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, messageId, ip, ua, subid, userId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func deleteDeletedUsers(prefix string, userIdHash string) {
	_, err := env.db.Exec("DELETE FROM deleted_users"+prefix+
		"WHERE id_user_md5 = ? ",
		userIdHash)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertDeny(coutry string, city string, blockId string, siteId string, ip string,
	ua string, subid string, trafficOwner string) {
	_, err := env.db.Exec("INSERT INTO log_deny ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner "+
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?)",
		coutry, city, siteId, blockId, ip, ua, subid, trafficOwner)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertPop(coutry string, city string, blockId string, siteId string, ip string,
	ua string, subid string, trafficOwner string) {
	_, err := env.db.Exec("INSERT INTO log_popup ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner "+
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?)",
		coutry, city, siteId, blockId, ip, ua, subid, trafficOwner)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertUniquePop(coutry string, city string, blockId string, siteId string, ip string,
	ua string, subid string, trafficOwner string) {
	_, err := env.db.Exec("INSERT INTO log_popup_uniq ( "+
		"date, "+
		"time, "+
		"country_code, "+
		"city_code, "+
		"id_site, "+
		"id_block, "+
		"ip, "+
		"user_agent, "+
		"subid, "+
		"traffic_owner "+
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?)",
		coutry, city, siteId, blockId, ip, ua, subid, trafficOwner)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertClicksShaving(coutry string, city string, block *Block, ip string, ua string, subid string, messageId string, userId string) {

	_, err := env.db.Exec("INSERT INTO log_clicks_shaving ( "+
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
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, messageId, ip, ua, subid, userId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}

func insertClicks(coutry string, city string, block *Block, ip string, ua string, subid string, messageId string, userId string) {
	_, err := env.db.Exec("INSERT INTO log_clicks ( "+
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
		") VALUES (CURDATE(), CURTIME(), ?,?,?,?,?,?,?,?,?)",
		coutry, city, block.IdSite, block.IdBlock, messageId, ip, ua, subid, userId)

	if err != nil {
		Error.Print(err.Error())
		return
	}
	return
}
