#!/bin/bash
PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin


# curl -fsSL  https://raw.githubusercontent.com/midoks/tgbot/master/scripts/install_dev.sh | sh

# Linux 手动安装
# wget https://go.dev/dl/go1.19.1.linux-amd64.tar.gz
# sudo tar -C /usr/local -xzf go1.19.1.linux-amd64.tar.gz
# sudo ln -s /usr/local/go/bin/* /usr/bin/

# systemctl status tgbot

# 手动编译
# go build main.go -o tgbot && tgbot web 

# Debug Now
export PATH=/usr/local/go:$PATH:/root/go/bin
export GOPATH=/root/go


TAGRT_DIR=/usr/local/tgbot_dev
mkdir -p $TAGRT_DIR
cd $TAGRT_DIR

export GIT_COMMIT=$(git rev-parse HEAD)
export BUILD_TIME=$(date -u '+%Y-%m-%d %I:%M:%S %Z')

go install github.com/midoks/zzz@latest

if [ ! -d $TAGRT_DIR/tgbot ]; then
	git clone https://github.com/midoks/tgbot
	cd $TAGRT_DIR/tgbot
else
	cd $TAGRT_DIR/tgbot
	git pull
fi

go mod tidy
go mod vendor

# cd /usr/local/tgbot_dev/tgbot && go build -o tgbot main.go 
# cd /usr/local/tgbot_dev/tgbot && go build -o tgbot main.go && ./tgbot web
cd $TAGRT_DIR/tgbot && go build -o tgbot main.go 
systemctl daemon-reload


cd $TAGRT_DIR/tgbot && ./tgbot install
systemctl restart tgbot


# rm -rf /usr/local/tgbot_dev/tgbot/custom
# rm -rf /usr/local/tgbot_dev/tgbot/data

cd $TAGRT_DIR/tgbot && ./tgbot -v

if [ ! -d /usr/local/go ];then
	wget https://golang.google.cn/dl/go1.26.2.linux-amd64.tar.gz
	tar -xvf go1.26.2.linux-amd64.tar.gz
	mv go /usr/local/
fi


if [ ! -f /root/go/bin/zzz ];then
	go install github.com/midoks/zzz@latest
fi

systemctl status tgbot

