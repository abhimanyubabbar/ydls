package iso639

// ShortToLong map 2 letter to 3 letter language code
// Based on https://www.loc.gov/standards/iso639-2/ISO-639-2_utf-8.txt
// Generated with:
// cat ISO-639-2_utf-8.txt | awk '-F|' '{ printf "\"%s\": \"%s\",\n",$3,$1 }' | grep -v '^"":'
var ShortToLong = map[string]string{
	"aa": "aar",
	"ab": "abk",
	"af": "afr",
	"ak": "aka",
	"sq": "alb",
	"am": "amh",
	"ar": "ara",
	"an": "arg",
	"hy": "arm",
	"as": "asm",
	"av": "ava",
	"ae": "ave",
	"ay": "aym",
	"az": "aze",
	"ba": "bak",
	"bm": "bam",
	"eu": "baq",
	"be": "bel",
	"bn": "ben",
	"bh": "bih",
	"bi": "bis",
	"bs": "bos",
	"br": "bre",
	"bg": "bul",
	"my": "bur",
	"ca": "cat",
	"ch": "cha",
	"ce": "che",
	"zh": "chi",
	"cu": "chu",
	"cv": "chv",
	"kw": "cor",
	"co": "cos",
	"cr": "cre",
	"cs": "cze",
	"da": "dan",
	"dv": "div",
	"nl": "dut",
	"dz": "dzo",
	"en": "eng",
	"eo": "epo",
	"et": "est",
	"ee": "ewe",
	"fo": "fao",
	"fj": "fij",
	"fi": "fin",
	"fr": "fre",
	"fy": "fry",
	"ff": "ful",
	"ka": "geo",
	"de": "ger",
	"gd": "gla",
	"ga": "gle",
	"gl": "glg",
	"gv": "glv",
	"el": "gre",
	"gn": "grn",
	"gu": "guj",
	"ht": "hat",
	"ha": "hau",
	"he": "heb",
	"hz": "her",
	"hi": "hin",
	"ho": "hmo",
	"hr": "hrv",
	"hu": "hun",
	"ig": "ibo",
	"is": "ice",
	"io": "ido",
	"ii": "iii",
	"iu": "iku",
	"ie": "ile",
	"ia": "ina",
	"id": "ind",
	"ik": "ipk",
	"it": "ita",
	"jv": "jav",
	"ja": "jpn",
	"kl": "kal",
	"kn": "kan",
	"ks": "kas",
	"kr": "kau",
	"kk": "kaz",
	"km": "khm",
	"ki": "kik",
	"rw": "kin",
	"ky": "kir",
	"kv": "kom",
	"kg": "kon",
	"ko": "kor",
	"kj": "kua",
	"ku": "kur",
	"lo": "lao",
	"la": "lat",
	"lv": "lav",
	"li": "lim",
	"ln": "lin",
	"lt": "lit",
	"lb": "ltz",
	"lu": "lub",
	"lg": "lug",
	"mk": "mac",
	"mh": "mah",
	"ml": "mal",
	"mi": "mao",
	"mr": "mar",
	"ms": "may",
	"mg": "mlg",
	"mt": "mlt",
	"mn": "mon",
	"na": "nau",
	"nv": "nav",
	"nr": "nbl",
	"nd": "nde",
	"ng": "ndo",
	"ne": "nep",
	"nn": "nno",
	"nb": "nob",
	"no": "nor",
	"ny": "nya",
	"oc": "oci",
	"oj": "oji",
	"or": "ori",
	"om": "orm",
	"os": "oss",
	"pa": "pan",
	"fa": "per",
	"pi": "pli",
	"pl": "pol",
	"pt": "por",
	"ps": "pus",
	"qu": "que",
	"rm": "roh",
	"ro": "rum",
	"rn": "run",
	"ru": "rus",
	"sg": "sag",
	"sa": "san",
	"si": "sin",
	"sk": "slo",
	"sl": "slv",
	"se": "sme",
	"sm": "smo",
	"sn": "sna",
	"sd": "snd",
	"so": "som",
	"st": "sot",
	"es": "spa",
	"sc": "srd",
	"sr": "srp",
	"ss": "ssw",
	"su": "sun",
	"sw": "swa",
	"sv": "swe",
	"ty": "tah",
	"ta": "tam",
	"tt": "tat",
	"te": "tel",
	"tg": "tgk",
	"tl": "tgl",
	"th": "tha",
	"bo": "tib",
	"ti": "tir",
	"to": "ton",
	"tn": "tsn",
	"ts": "tso",
	"tk": "tuk",
	"tr": "tur",
	"tw": "twi",
	"ug": "uig",
	"uk": "ukr",
	"ur": "urd",
	"uz": "uzb",
	"ve": "ven",
	"vi": "vie",
	"vo": "vol",
	"cy": "wel",
	"wa": "wln",
	"wo": "wol",
	"xh": "xho",
	"yi": "yid",
	"yo": "yor",
	"za": "zha",
	"zu": "zul",
}
