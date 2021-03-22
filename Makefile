hello:
	echo "ifundrobot"

build:
	go build -o output/ifundrobot-server server/main.go
	go build -o output/ifundrobot-client client/main.go

all: hello build

clean:
	rm -f output/*