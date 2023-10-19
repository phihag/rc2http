
default: rc2http rc2http-host

rc2http: *.go static/
	GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o rc2http

rc2http-host: *.go static/
	go build -o rc2http-host

clean:
	rm -f rc2http rc2http-host

.PHONY: clean
