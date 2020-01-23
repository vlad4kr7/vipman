DEFAULT_GOAL := build

build :
	 cd src && go test dbinvent.com/vipman/vman && go build -o ../bin/vipman main/main.go 

build-echo :
	 cd src && go build -o ../bin/wecho sample/echo-sample.go

build-win-echo :
	 echo "go build -o bin/wecho.exe src/sample/echo-sample.go"
