
build:
	go build -ldflags "-X 'main.mainVersion=v0.0.1' -X 'main.buildVersion=${pwd}'"