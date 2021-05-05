hello:
	echo "ifundrobot"

build:
	env GOOS=linux GOARCH=amd64 go build -o output/ifundrobot-server server/main.go
	env GOOS=linux GOARCH=amd64 go build -o output/ifundrobot-client client/main.go

all: hello build

clean:
	rm -f output/*