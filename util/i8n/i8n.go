package i8n

var i8n = make(map[string]map[int]string)

const Default = "zh-CN"

var local = Default

func GetString(key int, lang ...string) string {
	if len(lang) != 0 {
		return i8n[lang[0]][key]
	}
	if local != "" {
		return i8n[local][key]
	}
	return i8n[Default][key]
}
