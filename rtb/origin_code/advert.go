package main

type Block struct {
	IdBlock       string
	IdSite        string
	PayType       string
	Shaving       int
	MessageString string
	UniqeShowsRtb int
}

type Subid struct {
	IdBlock string
	SubId   string
}

type Message struct {
	Id                 string
	Url                string
	Title              string
	Body               string
	Image              string
	Icon               string
	Image_url          string
	Icon_url           string
	Content_type       string
	Datetime_available string
	Message_priority   string
	Dir                string
	Show_period        int
}

type CountryCodes struct {
	Code string
}

type Cpc struct {
	Cpc float64
}

//General data to response

type Geo struct {
	Country string `json:"country"`
}
type Site struct {
	Id string `json:"id,omitempty"`
}

type Device struct {
	Ip  string `json:"ip"`
	Ua  string `json:"ua"`
	Geo Geo    `json:"geo,omitempty"`
}

type Link struct {
	Url string `json:"url"`
}

type Title struct {
	Text string `json:"text"`
}

type AssetTitle struct {
	Id       int   `json:"id"`
	Required int   `json:"required"`
	Title    Title `json:"title"`
}
type AssetTitleNotRequired struct {
	Id    int   `json:"id"`
	Title Title `json:"title"`
}
type Value struct {
	Value string `json:"value"`
}
type AssetText struct {
	Id       int   `json:"id"`
	Required int   `json:"required"`
	Data     Value `json:"data"`
}
type AssetTextNotRequired struct {
	Id   int   `json:"id"`
	Data Value `json:"data"`
}

type Img struct {
	W   int    `json:"w"`
	H   int    `json:"h"`
	Url string `json:"url"`
}
type AssetImage struct {
	Id       int `json:"id"`
	Required int `json:"required"`
	Img      Img `json:"img"`
}

type AssetImageNotRequired struct {
	Id  int `json:"id"`
	Img Img `json:"img"`
}

type Ext struct {
	Ctr float64 `json:"ctr"`
}

func (env *Env) GetCountryByCode(cc string) error {
	countryCode := &CountryCodes{}

	err := env.db.QueryRow(
		"SELECT code "+
			"FROM country_codes "+
			"WHERE code = ?", cc).Scan(&countryCode.Code)

	if err != nil {
		return err
	}

	return nil
}

func (env *Env) GetBlockById(key string) (*Block, error) {
	block := &Block{}

	err := env.db.QueryRow(
		"SELECT id_block, id_site, pay_type, shaving, message_string, unique_shows_rtb "+
			"FROM message_blocks "+
			"WHERE id_block = ?", key).Scan(&block.IdBlock, &block.IdSite, &block.PayType,
		&block.Shaving, &block.MessageString, &block.UniqeShowsRtb)

	if err != nil {
		return nil, err
	}

	return block, nil
}

func (env *Env) GetCpc(where string) (*Cpc, error) {
	cpc := &Cpc{}

	err := env.db.QueryRow(
		"SELECT cpc FROM wm_tariffs_message " +
			"WHERE " + where +
			"ORDER BY id_block DESC LIMIT 1").Scan(&cpc.Cpc)

	if err != nil {
		return cpc, err
	}

	return cpc, nil
}

func (env *Env) GetMessages(where string) (Message, error) {
	results, err := env.db.Query(
		"SELECT " +
			"id, url, title, body, image, icon, image_url, " +
			"icon_url, content_type, datetime_available, message_priority, " +
			"show_period, 'ltr' dir " +
			"FROM advert_message " +
			"WHERE " +
			where +
			"ORDER BY message_priority ASC")
	defer results.Close()
	message := Message{}
	if err != nil {
		return message, err
	}

	messagesByPriority := []Message{}
	lowestPriority := ``
	for results.Next() {
		// for each row, scan the result into our tag composite object
		err = results.Scan(&message.Id, &message.Url, &message.Title, &message.Body, &message.Image,
			&message.Icon, &message.Image_url, &message.Icon_url,
			&message.Content_type, &message.Datetime_available, &message.Message_priority, &message.Show_period,
			&message.Dir)

		if err != nil {
			return message, err
		}
		if len(lowestPriority) == 0 || lowestPriority == message.Message_priority {
			lowestPriority = message.Message_priority
			messagesByPriority = append(messagesByPriority, message)
		}

	}

	messages := messagesByPriority
	messagesByPriority = nil

	messageResult := messages[RandIntMapKey(messages)]
	messages = nil
	return messageResult, nil
}

func (env *Env) BlockBlackListSubidByBlockAndSubid(key string, subid string) bool {
	subidObj := &Subid{}

	err := env.db.QueryRow(
		"SELECT id_block, id_subid "+
			"FROM blacklist_block_subid "+
			"WHERE id_block = ? AND id_subid = ?", key, subid).Scan(&subidObj.IdBlock, &subidObj.SubId)

	if err != nil {
		return false
	}

	return true
}

func (env *Env) BlockWhiteListSubidByBlockAndSubid(key string, subid string) bool {
	results, err := env.db.Query(
		"SELECT id_block, id_subid "+
			"FROM whitelist_block_subid "+
			"WHERE id_block = ?", key)

	if err != nil {
		return false
	}
	defer results.Close()
	subidObj := &Subid{}
	anyData := false
	for results.Next() {
		// for each row, scan the result into our tag composite object
		err := results.Scan(&subidObj.IdBlock, &subidObj.SubId)

		if err != nil {
			return false
		}
		anyData = true
		if subidObj.SubId == subid {
			return false
		}

	}

	return anyData
}
