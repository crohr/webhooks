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
		mkdir -p package/etc/profile.d && mkdir -p package/etc/init;\
	fi
	cp build/webhook package/usr/local/bin/
	cp config.yaml.sample package/etc/webhook/
	cp webhook.sh.sample package/etc/profile.d/webhook.sh
	cp webhook.conf package/etc/init/
