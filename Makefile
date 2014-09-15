default: build

build: clean folder
	go build -o bin/webhook push/push.go

folder:
	mkdir -p bin

clean:
	if [ -d "bin" ]; then rm -rf bin; fi

