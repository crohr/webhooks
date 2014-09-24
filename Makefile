default: build

build: clean folder
	go build -o build/webhook push/push.go

folder:
	if ! [ -d "build" ]; then mkdir -p build; fi

clean:
	if [ -d "build" ] && [ -e "build/webhook" ]; then rm build/webhook*; fi

rpm: default
	if [ -d "package" ]; then  rm -rf package && mkdir package; fi
	cp build/webhook package/
	cp config.yaml.sample package/config.yaml
	cp webhook.sh.sample package/webhook.sh
	cp webhook.conf package/
	fpm -t rpm -s dir -n webhook -p build -a x86_64 \
		-m "biosidd@gmail.com"  -v "1.0.0" \
		--config-files /etc/webhook/config.yaml\
		./package/webhook=/usr/local/bin/webhook \
		./package/webhook.sh=/etc/profile.d/webhook.sh \
		./package/config.yaml=/etc/webhook/config.yaml \
		./package/webhook.conf=/etc/init/webhook.conf

