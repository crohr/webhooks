default: build

build: clean folder
	go build -o build/webhook push/push.go

folder:
	if ![ -d "build" ]; then mkdir -p build; fi

clean:
	if [ -d "build" ] && [ -e "build/webhook" ]; then rm build/webhook; fi

rpm: default
	if ! [ -d "package" ]; then \
		rm -rf package; \
		mkdir -p package/usr/local/bin && mkdir -p package/etc/webhook;\
	fi
	cp build/webhook package/usr/local/bin/
	cp config.yaml.sample package/etc/webhook/

