package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/mssola/user_agent"
	"math/rand"
	"strings"
)

var (
	tbl_prefix = map[int64]string{
		99436580200:  "",
		195065829525: "_195065829525",
	}
	user_prefix = map[int64]string{
		99436580200:  "_99436580200",
		195065829525: "_195065829525",
	}
	brokers = [9]string{
		"azartzona",
		"bigazart",
		"golden",
		"grand",
		"igrun",
		"rubet",
		"xcasino",
		"zeon",
		"zonaigr",
	}
	letters = []rune("0123456789")
	devices = [5]string{
		`Android`,
		`Phone`,
		`iPad`,
		`iPod`,
	}
	deviceToUrls = map[string]string{
		"pc":      "0",
		"android": "1",
		"iphone":  "2",
		"ipad":    "3",
		"ipod":    "4",
		"zz":      "5",
	}
	countriesWithCity = map[string]string{
		"RU": "RU",
		"UA": "UA",
		"KZ": "KZ",
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
	citiesLike = map[string]map[string]string{
		"RU": {
			"chulim":                   "Chulym",
			"tselina":                  "Zelina",
			"tayga":                    "Taiga",
			"solnechnogorskiy":         "Solnechnogorsk",
			"orekhovo":                 "Orekhovo-zuevo",
			"nyazepetrovsk":            "Nyazepetrovskaya",
			"nizhnekamskiy":            "Nizhnekamsk",
			"nakhodka":                 "Nahodka",
			"leninsk-kuznetsky":        "Leninsk-kuznetskiy",
			"kotelnikowo":              "Kotelnikovo",
			"kopeysk":                  "Kopeisk",
			"kislovodskaya":            "Kislovodsk",
			"belv":                     "Belovo",
			"chepetsk":                 "Kirovo-chepetsk",
			"mikhaylovskaya":           "Sochi",
			"severoural'sk":            "Severouralsk",
			"rel'":                     "Rel",
			"privol'noye":              "Privolnoye",
			"buzov'yazy":               "Buzovyazy",
			"bol'sherech'ye":           "Bolsherechye",
			"krasnotur'insk":           "Krasnoturinsk",
			"gus-khrustalnyi":          "Gus-khrustalnyy",
			"krasnyy sulin":            "Sulin",
			"mogoitui":                 "Mogoytuy",
			"gur'yevsk":                "Guryevsk",
			"arsen'yev":                "Arsenyev",
			"shlissel'burg":            "Shlisselburg",
			"firsanovka":               "Khimki",
			"mosrentgen":               "Moscow",
			"skolkovo":                 "Moscow",
			"aksai":                    "Aksay",
			"van'kino":                 "Vankino",
			"velikiye luki":            "Velikie Luki",
			"gubkinsky":                "Gubkinskiy",
			"kirowsk":                  "Kirovsk",
			"nijni novgorod":           "Nizhni Novgorod",
			"nishnij nowgorod":         "Nizhni Novgorod",
			"nizhnaya tura":            "Nizhnyaya Tura",
			"petrovskii":               "Petrovski",
			"svobodny":                 "Svobodnyy",
			"uspenskoye":               "Uspenskoe",
			"engels":                   "Engels-Yurt",
			"engel's":                  "Engels",
			"moskva":                   "Moscow",
			"moskwa":                   "Moscow",
			"alexandrov":               "Aleksandrov",
			"arkhangelsk":              "Archangelsk",
			"bataysk":                  "Bataisk",
			"yekaterinburg":            "Sverdlovsk",
			"zheleznodorozhnyy":        "Zheleznodorozhny",
			"mozhaysk":                 "Mozhaisk",
			"mytishi":                  "Mytishchi",
			"nizhniy novgorod":         "Nizhni Novgorod",
			"nizhniy tagil":            "Nizhnii Tagil",
			"novorossisk":              "Novorossiisk",
			"novorossiysk":             "Novorossiisk",
			"rostov-on-don":            "Rostov-na-donu",
			"saint petersburg":         "Leningrad",
			"slavyansk-na-kubani":      "Slavyansk",
			"snezhinsk":                "Chelyabinsk-70",
			"tolyatti":                 "Togliatti",
			"tuymazy":                  "Tuimazy",
			"chelyabinsk":              "Cheliabinsk",
			"yeysk":                    "Yeisk",
			"alekseyevka":              "Alekseevka",
			"al'met'yevsk":             "Almetyevsk",
			"belebei":                  "Belebey",
			"bodaybo":                  "Bodaibo",
			"buinaksk":                 "Buynaksk",
			"gelendjik":                "Gelendzhik",
			"ekaterinburg":             "Sverdlovsk",
			"elabuga":                  "Yelabuga",
			"yeniseisk":                "Eniseisk",
			"zaraisk":                  "Zaraysk",
			"zarechnyy":                "Zarechny",
			"krasnojarsk":              "Krasnoyarsk",
			"krasny-sulin":             "Sulin",
			"leninsk-kuznetsk":         "Kuznetsk",
			"lodeynoye pole":           "Lodeyno",
			"lubertsi":                 "Lyubertsy",
			"maikop":                   "Maykop",
			"mineralnye":               "Mineralnyye Vody",
			"mirnyy":                   "Mirny",
			"mzensk":                   "Mtsensk",
			"naberezhnyye chelny":      "Chelny",
			"nizhnii novgorod":         "Nizhni Novgorod",
			"nijnii novgorod":          "Nizhni Novgorod",
			"ob'":                      "Ob",
			"orekhovo-zuyevo":          "Orekhovo-zuevo",
			"otradnoye":                "Otradnoe",
			"pervoural'sk":             "Pervouralsk",
			"poslok":                   "Poselok",
			"prokop'yevsk":             "Prokopyevsk",
			"ramenskoe":                "Ramenskoye",
			"ryazan'":                  "Ryazan",
			"sankt-peterburg":          "Leningrad",
			"slyudyanka":               "Sludyanka",
			"staraya ruza":             "Staraya Russa",
			"stary oskol":              "Staryy Oskol",
			"tol'yatti":                "Togliatti",
			"tjumen":                   "Tyumen",
			"urjupinsk":                "Uryupinsk",
			"usol'ye-sibirskoye":       "Usolye-sibirskoye",
			"ust'-ilimsk":              "Ust-ilimsk",
			"schadrinsk":               "Shadrinsk",
			"schachty":                 "Shakhty",
			"artmovskiy":               "Artemovski",
			"novaya balakhna":          "Balakhna",
			"gus-khrustalny":           "Gus-khrustalnyy",
			"izobil'nyy":               "Izobilnyy",
			"korolv":                   "Korolyov",
			"medvezhegorsk":            "Medvezhyegorsk",
			"otradny":                  "Otradnyy",
			"peterhof":                 "Petergof",
			"pitkäranta":               "Pitkyaranta",
			"solnechnodol'sk":          "Solnechnodolsk",
			"ozry":                     "Ozyory",
			"orl":                      "Oryol",
			"vyshniy volochk":          "Vyshniy Volochyok",
			"ozrsk":                    "Ozyorsk",
			"ogarv":                    "Ogaryov",
			"pugachvo":                 "Pugachyovo",
			"shchkino":                 "Shchyokino",
			"venv":                     "Venyov",
			"grachvka":                 "Grachyovka",
			"budnnovsk":                "Budyonnovsk",
			"berzovskiy":               "Beryozovskiy",
			"timashvo":                 "Timashyovo",
			"artm":                     "Artyom",
			"ochr":                     "Ochyor",
			"moskalnki":                "Moskalyonki",
			"linvo":                    "Linyovo",
			"fdorovskoye":              "Fyodorovskoye",
			"semnov":                   "Semyonov",
			"alshkovo":                 "Alyoshkovo",
			"zaozrsk":                  "Zaozyorsk",
			"shchlkovo":                "Shchyolkovo",
			"shlkovo":                  "Shyolkovo",
			"mikhnvo":                  "Mikhnyovo",
			"priozrsk":                 "Priozyorsk",
			"pikalvo":                  "Pikalyovo",
			"vichvshchina":             "Vichyovshchina",
			"kiselvsk":                 "Kiselyovsk",
			"alkhino":                  "Alyokhino",
			"semnovskoye":              "Semyonovskoye",
			"gusinoozrsk":              "Gusinoozyorsk",
			"pochp":                    "Pochyop",
			"beloozrskiy":              "Beloozyorskiy",
			"beloozлrskiy":             "Beloozyorskiy",
			"belorezk":                 "Beloretsk",
			"biisk":                    "Biysk",
			"verkhniy ufaley":          "Ufaley",
			"vsevolozhskiy":            "Vsevolozhsk",
			"viborg":                   "Vyborg",
			"vylgort":                  "Vilgort",
			"gai":                      "Gay",
			"georgiyevsk":              "Georgievsk",
			"dal'negorsk":              "Dalnegorsk",
			"ivanovskaya":              "Ivanovo",
			"kamen":                    "Kamen-na-obi",
			"kirow":                    "Kirov",
			"belaya":                   "Kirovo-chepetsk",
			"koryazma":                 "Koryazhma",
			"kurovskoy":                "Kurovskoye",
			"lipezk":                   "Lipetsk",
			"gorki":                    "Nizhni Novgorod",
			"avtozavodskiy rayon":      "Nizhni Novgorod",
			"dubroye":                  "Obninsk",
			"irtysh":                   "Omsk",
			"zalesskiy":                "Pereslavl-zalesskiy",
			"permskiy":                 "Perm",
			"pitkranta":                "Pitkyaranta",
			"russa":                    "",
			"khabarovskaya":            "Khabarovsk",
			"aibga":                    "Sochi",
			"poslok vorotynsk":         "Vorotynsk",
			"zelenoborskoye":           "Zelenoborskiy",
			"altayskoye":               "Altaiskoe",
			"arkhangel'sk":             "Archangelsk",
			"beloozerskiy":             "Beloozyorskiy",
			"berezovskiy":              "Beryozovskiy",
			"budennovsk":               "Budyonnovsk",
			"gornoural'skiy":           "Gornouralskiy",
			"groznyy":                  "Grozny",
			"kamensk-ural'skiy":        "Kamensk-uralskiy",
			"kiselevsk":                "Kiselyovsk",
			"kol'chugino":              "Kolchugino",
			"komsomolsk-on-amur":       "Komsomolsk-na-amure",
			"korolev":                  "Korolyov",
			"kotel'niki":               "Kotelniki",
			"krasnoural'sk":            "Krasnouralsk",
			"mineralnye vody":          "Mineralnyye Vody",
			"monetnyy":                 "Monetny",
			"nal'chik":                 "Nalchik",
			"nizhny tagil":             "Nizhnii Tagil",
			"oktyabr'skiy":             "Oktyabrskiy",
			"orel":                     "Oryol",
			"ozersk":                   "Ozyorsk",
			"pereslavl'-zalesskiy":     "Pereslavl-zalesskiy",
			"petropavlovsk-kamchatsky": "Petropavlovsk-kamchatskiy",
			"roslavl'":                 "Roslavl",
			"staroshcherbinovskaya":    "Starosherbinovskaya",
			"stavropol'":               "Stavropol",
			"syzran'":                  "Syzran",
			"trekhgornyy":              "Trekhgorny",
			"ust'-katav":               "Ust-katav",
			"ust'-labinsk":             "Ust-labinsk",
			"vel'sk":                   "Velsk",
			"vyshniy volochek":         "Vyshniy Volochyok",
			"yuryuzan'":                "Yuryuzan",
			"start oskol":              "Staryy Oskol",
		},
		"UA": {
			"khmelnytskyi":           "Khmelnytskyy",
			"fastow":                 "Fastov",
			"snyatyn":                "Snyatin",
			"pryluky":                "Priluki",
			"marhanets":              "Marganets",
			"makeyevka":              "Makiyivka",
			"lozova":                 "Lozovaya",
			"ladyzhin":               "Ladyzhyn",
			"kurakhove":              "Kurakhovo",
			"sorokino":               "Krasnodon",
			"kirovskoye":             "Kirovske",
			"zytomierz":              "Zhytomyr",
			"volochysk":              "Volochisk",
			"berezhani":              "Berezhany",
			"bakhmut":                "Artemovsk",
			"ivano-frankivs'k":       "Ivanofrankovsk",
			"ivanofrankivsk":         "Ivanofrankovsk",
			"khartsyz'k":             "Khartsyzk",
			"l'viv":                  "Lvov",
			"luck":                   "Lutsk",
			"bila":                   "Bila Tserkva",
			"marhanets'":             "Marhanets",
			"dnepropetrowsk":         "Dnipropetrovsk",
			"kalynivka":              "Kalinovka",
			"krasnyy luch":           "Krasny Luch",
			"nizhin":                 "Nizhyn",
			"chasiv yar":             "Chasov Yar",
			"kiyiv":                  "Kiev",
			"belaya tserkov":         "Bila Tserkva",
			"berdichev":              "Berdychiv",
			"borispol":               "Boryspil",
			"dneprodzerzhinsk":       "Dniprodzerzhynsk",
			"dniepropetrovsk":        "Dnipropetrovsk",
			"dnepropetrovsk":         "Dnipropetrovsk",
			"donets":                 "Donetsk",
			"zhitomir":               "Zhytomyr",
			"zaporozhe":              "Zaporozhye",
			"zaporizhzhya":           "Zaporozhye",
			"ilichevsk":              "Illichivsk",
			"kirovograd":             "Kirovohrad",
			"kremenchug":             "Kremenchuk",
			"lugansk":                "Luhansk",
			"lviv":                   "Lvov",
			"makeevka":               "Makiyivka",
			"nezhin":                 "Nizhyn",
			"netishin":               "Netishyn",
			"mykolayiv":              "Nikolaev",
			"obukhiv":                "Obukhov",
			"rivne":                  "Rovno",
			"sebastopol":             "Sevastopol",
			"smela":                  "Smila",
			"sumi":                   "Sumy",
			"ternopil":               "Ternopol",
			"tarnopol":               "Ternopol",
			"uzhgorod":               "Uzhhorod",
			"fastiv":                 "Fastov",
			"kharkiv":                "Kharkov",
			"charkow":                "Kharkov",
			"khmelnitskiy":           "Khmelnytskyy",
			"cherkassy":              "Cherkasy",
			"chernigov":              "Chernihiv",
			"chernivtsi":             "Chernovtsy",
			"tal'ne":                 "Talne",
			"talne":                  "Talne",
			"oleksandriya":           "Aleksandriya",
			"antratsyt":              "Antratsit",
			"bakhchisaray":           "Bakhchysaray",
			"vasylkiv":               "Vasilkov",
			"vinnytsya":              "Vinnitsa",
			"vyshgorod":              "Vyshhorod",
			"dymytrov":               "Dimitrov",
			"drohobych":              "Drogobych",
			"zugres":                 "Zuhres",
			"irpen":                  "Irpin",
			"kazatin":                "Kozyatyn",
			"kanev":                  "Kaniv",
			"kyiv":                   "Kiev",
			"kyyiv":                  "Kiev",
			"kiew":                   "Kiev",
			"korostyshev":            "Korostyshiv",
			"krasnograd":             "Krasnohrad",
			"kryvyi rih":             "Krivoy Rog",
			"letichev":               "Letychiv",
			"lutuhyne":               "Lutugino",
			"pavlohrad":              "Pavlograd",
			"pervomaisk":             "Pervomaysk",
			"pervomays'k":            "Pervomaysk",
			"pereyaslavkhmelnytskyy": "Pereyaslav",
			"svitlovodsk":            "Svetlovodsk",
			"syevyerodonets'k":       "Severodonetsk",
			"sloviansk":              "Slavyansk",
			"slovyansk":              "Slavyansk",
			"vuhledar":               "Ugledar",
			"cherson":                "Kherson",
			"shepetovka":             "Shepetivka",
			"yasynuvata":             "Yasinovataya",
			"balakliya":              "Balakleya",
			"baryshivka":             "Baryshevka",
			"artemivsk":              "Artemovsk",
			"bohorodchany":           "Bogorodchany",
			"izmayil":                "Izmail",
			"mukacheve":              "Mukachevo",
			"rozhishche":             "Rozhyshche",
			"starokostyantyniv":      "Starokonstantinov",
			"stryy":                  "Stryi",
			"uzyn":                   "Uzin",
			"feodosia":               "Feodosiya",
			"alchevs'k":              "Alchevsk",
			"belaya":                 "Bila Tserkva",
			"derhachi":               "Dergachi",
			"kvasyliv":               "Kvasilov",
			"krasyliv":               "Krasilov",
			"lwow":                   "Lvov",
			"odesa":                  "Odessa",
			"tyachev":                "Tyachiv",
			"ungvar":                 "Uzhhorod",
			"chervonohrad":           "Chervonograd",
			"ivano-frankivsk":        "Ivanofrankovsk",
			"vinnytsia":              "Vinnitsa",
			"kryvyy rih":             "Krivoy Rog",
			"zdolbunov":              "Zdolbuniv",
		},
		"KZ": {
			"zhansgirov":       "Taldykorgan",
			"qyzylorda":        "Kyzylorda",
			"kapchagay":        "Kapchagai",
			"ziryanovsk":       "Zyryanovsk",
			"dzhezkazgan":      "Zhezkazgan",
			"ermentau":         "Ereymentau",
			"aqtau":            "Aktau",
			"aktobe":           "Aktubinsk",
			"aqtbe":            "Aktubinsk",
			"karagandy":        "Karaganda",
			"qostanay":         "Kostanay",
			"petropavl":        "Petropavlovsk",
			"zhambyl":          "Taraz",
			"skemen":           "Ust-kamenogorsk",
			"ust'-kamenogorsk": "Ust-kamenogorsk",
			"shevchenko":       "Aktau",
			"vostok":           "UN",
			"saran'":           "Saran",
			"semipalatinsk":    "Semey",
			"trgen":            "Turgen",
			"aktyubinsk":       "Akt",
		},
	}
)

func GetUserCity(country_code string, city_code string) string {
	if _, ok := countriesWithCity[country_code]; ok {
		if len(city_code) == 0 {
			return "UN"
		}
		city_code = strings.ReplaceAll(city_code, "\x00-", "")
		city_code = strings.ReplaceAll(city_code, "\x1F", "")
		city_code = strings.ReplaceAll(city_code, "\x80-", "")
		city_code = strings.ReplaceAll(city_code, "\xFF", "")
		city_code = strings.ReplaceAll(city_code, "\xF6", "")
		city_code = strings.ReplaceAll(city_code, "\xD6", "")

		return city_code

	}
	return "UN"
}

func IsMobile(ua string) bool {
	uaObj := user_agent.New(ua)
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

func RandSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
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
