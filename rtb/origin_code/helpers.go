package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/mssola/user_agent"
	"math/rand"
	"net/url"
	"strings"
)

var (
	devices = [5]string{
		`Android`,
		`iPhone`,
		`iPad`,
		`iPod`,
	}
	supportedPcBrowsers = map[string]string{
		"Firefox":   "Firefox",
		"Opera":     "Opera",
		"Chrome":    "Chrome",
		"MSIE":      "MSIE",
		"Safari":    "Safari",
		"YaBrowser": "YaBrowser",
		"Tv":        "Tv",
	}
	browserGroups = map[string]string{
		"ie1":         "MSIE",
		"ie>1":        "MSIE",
		"ie11":        "MSIE",
		"ie_edge":     "MSIE",
		"firefox":     "Firefox",
		"opera15plus": "Opera",
		"opera":       "Opera",
		"chrome":      "Chrome",
		"yandex":      "YaBrowser",
		"yabrowser":   "YaBrowser",
		"safari":      "Safari",
		"safariwin":   "Safari",
	}
	browsers = map[string]map[string]bool{
		"ie1": {
			"microsoft internet explorer": true,
		},
		"ie>1": {
			"opera":          false,
			"opr":            false,
			"msie":           true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"ie11": {
			"trident/7.0":    true,
			"rv:11.0":        true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"ie_edge": {
			"yabrowser":      false,
			"opr":            false,
			"midori":         false,
			"version":        false,
			"chrome":         true,
			"edge":           true,
			"safari":         true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"safari": {
			"yabrowser":      false,
			"opr":            false,
			"midori":         false,
			"version":        true,
			"chrome":         true,
			"safari":         true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"safariwin": {
			"applewebkit":    true,
			"version":        true,
			"safari":         true,
			"chrome":         false,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"yabrowser": {
			"yabrowser":      true,
			"safari":         true,
			"chrome":         true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"yandex": {
			"yabrowser":      false,
			"yandex":         true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"opera15plus": {
			"opr":            true,
			"safari":         true,
			"chrome":         true,
			"yabrowser":      false,
			"midori":         false,
			"presto":         false,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"opera": {
			"opr":            false,
			"presto":         true,
			"opera":          true,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"chrome": {
			"yabrowser":      false,
			"opr":            false,
			"midori":         false,
			"version":        false,
			"chrome":         true,
			"safari":         true,
			"amigo":          false,
			"vivaldi":        false,
			"360browser":     false,
			"ucbrowser":      false,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
		"firefox": {
			"firefox":        true,
			") gecko/":       true,
			"opr":            false,
			"midori":         false,
			"version":        false,
			"chrome":         false,
			"safari":         false,
			"smart":          false,
			"freebsd":        false,
			"googletv":       false,
			"itunes-appletv": false,
			"tuner":          false,
		},
	}

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

func IsMobile(ua string) bool {
	newUa, _ := url.QueryUnescape(ua)
	uaObj := user_agent.New(newUa)
	return uaObj.Mobile()
}

func GetDevice(ua string) string {
	userAgent := strings.ToLower(ua)
	definedDeviceName := ``
	earliestPosition := -1
	intPosition := 0
	for _, value := range devices {
		intPosition = strings.Index(userAgent, strings.ToLower(value))
		if intPosition > -1 && intPosition > earliestPosition {
			earliestPosition = intPosition
			definedDeviceName = value
		}
	}
	if len(definedDeviceName) > 0 {
		return definedDeviceName
	} else {
		return `ZZ`
	}
}

func GetBrowserByUa(ua string) string {
	userAgent := strings.ToLower(ua)
	for browserName, pattern := range browsers {
		counter := 0
		for part, result := range pattern {
			if strings.Contains(userAgent, part) == result {
				counter++
			}
		}

		if counter == len(pattern) {
			if _, ok := supportedPcBrowsers[browserGroups[browserName]]; ok {
				return browserGroups[browserName]
			}
		}
	}

	return "ZZ"
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

func RandIntMapKey(m []Message) int {
	i := rand.Intn(len(m))
	for k := range m {
		if i == 0 {
			return k
		}
		i--
	}
	panic("never")
}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.Marshal(v)

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}
