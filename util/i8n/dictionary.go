package i8n

var Dictionary = new(DictionaryStruct)

type DictionaryStruct struct {
	CMD struct {
		Title string `json:"title"`
	} `json:"cmd"`
}
