package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/url"
)

var (
	countriesToMBCIS = map[string]string{
		`AZ`: `MBCIS`,
		`AM`: `MBCIS`,
		`GR`: `MBCIS`,
		`BY`: `MBCIS`,
		`KG`: `MBCIS`,
		`MD`: `MBCIS`,
		`TJ`: `MBCIS`,
		`TM`: `MBCIS`,
		`UZ`: `MBCIS`,
		`EE`: `MBCIS`,
		`LV`: `MBCIS`,
		`LT`: `MBCIS`,
	}

	countriesRUUAKZToMobile = map[string]string{
		`RU`: `MBR`,
		`UA`: `MBU`,
		`KZ`: `MBK`,
	}

	countriesToSNG = map[string]string{
		`AZ`: `SNG`,
		`AM`: `SNG`,
		`GR`: `SNG`,
		`KG`: `SNG`,
		`MD`: `SNG`,
		`TJ`: `SNG`,
		`TM`: `SNG`,
		`UZ`: `SNG`,
	}

	countriesToBAL = map[string]string{
		`EE`: `BAL`,
		`LV`: `BAL`,
		`LT`: `BAL`,
	}

	countiesNotInSpecialCountriesAndZz = map[string]string{
		`RU`: `RU`,
		`UA`: `UA`,
		`KZ`: `KZ`,
		`BY`: `BY`,
	}
)

func OpenSSLEncrypt(x string) string {
	iv, _ := base64.StdEncoding.DecodeString("AJf3QItKM7+Lkh/BZT2xNg==")
	key := []byte("1234567890abcdef1234567890abcdef")

	var plaintextblock []byte

	plaintext := x

	length := len(plaintext)
	if length%16 != 0 {
		extendBlock := 16 - (length % 16)
		plaintextblock = make([]byte, length+extendBlock)
		copy(plaintextblock[length:], bytes.Repeat([]byte{uint8(extendBlock)}, extendBlock))
	} else {
		plaintextblock = make([]byte, length)
	}
	copy(plaintextblock, plaintext)

	cipherBlock, err := aes.NewCipher(key)
	if err != nil {
		log.Fatal(err)
	}
	ciphertext := make([]byte, len(plaintextblock))
	mode := cipher.NewCBCEncrypter(cipherBlock, iv)
	mode.CryptBlocks(ciphertext, plaintextblock)
	str := url.QueryEscape(base64.StdEncoding.EncodeToString(ciphertext))

	return str
}
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func GetMobileGeoGroupCountry(country string) string {
	if _, ok := countriesToMBCIS[country]; ok {
		return countriesToMBCIS[country]
	} else if _, ok := countriesRUUAKZToMobile[country]; ok {
		return countriesRUUAKZToMobile[country]
	} else {
		return `MBZ`
	}
}

func GetPcGeoGroupCountry(country string) string {
	if _, ok := countriesToSNG[country]; ok {
		return countriesToSNG[country]
	} else if _, ok := countriesToBAL[country]; ok {
		return countriesToBAL[country]
	} else if _, ok := countiesNotInSpecialCountriesAndZz[country]; ok {
		return countiesNotInSpecialCountriesAndZz[country]
	} else {
		return `ZZ`
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func callToFcm(rb RequestBody, fcmKey string, fcmUrl string) (ResultsFcm, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fcmUrl)
	req.Header.SetMethod("POST")
	req.Header.Set("Authorization", "key="+fcmKey)
	req.Header.Set("Content-Type", "application/json")

	ResponseData := ResultsFcm{}
	js, errMarsh := json.Marshal(rb)
	if errMarsh != nil {
		return ResponseData, errMarsh
	}
	rb = RequestBody{}
	js = bytes.Replace(js, []byte("\\u003c"), []byte("<"), -1)
	js = bytes.Replace(js, []byte("\\u003e"), []byte(">"), -1)
	js = bytes.Replace(js, []byte("\\u0026"), []byte("&"), -1)
	req.SetBody(js)
	js = nil
	resp := fasthttp.AcquireResponse()
	err := env.clientHttp.Do(req, resp)
	defer resp.ConnectionClose()

	if err != nil {
		return ResponseData, err
	}
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	errDecode := json.Unmarshal(resp.Body(), &ResponseData)
	if errDecode != nil {
		return ResponseData, errDecode
	}

	return ResponseData, nil
}
