
build:
	go build -ldflags "-X 'main.mainVersion=v0.0.1' -X 'main.buildVersion=$(shell git rev-parse --short HEAD)'"

geni18n:
	go install github.com/Rehtt/Kit/i18n/geni18n@latest
	geni18n ./