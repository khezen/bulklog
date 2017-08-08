#!/bin/sh

echo "important files persistence..."
mkdir /default
cp /tmp/app/config.json /default/config.json
mv /tmp/app/entrypoint.sh /entrypoint.sh
chmod +x /entrypoint.sh
echo "done!"

echo "installing required packages..."
apk update
apk add ca-certificates
mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
apk add wget tar git
echo "done!"

echo "installing Go..."
GO_VERSION="1.8.1"
wget -O "/tmp/go$GO_VERSION.linux-amd64.tar.gz" "https://storage.googleapis.com/golang/go$GO_VERSION.linux-amd64.tar.gz"
tar -C /usr/local -xzf "/tmp/go$GO_VERSION.linux-amd64.tar.gz"
export PATH=$PATH:/usr/local/go/bin
echo "done"

echo "installing espipe..."
echo "dependencies..."
go get github.com/aws/aws-sdk-go/aws/signer/v4
go get github.com/aws/aws-sdk-go/aws/credentials
go get github.com/google/uuid
echo "compilation..."
mkdir -p /usr/local/go/src/github.com/khezen/
mv /tmp/app /usr/local/go/src/github.com/khezen/espipe
go build -o /bin/espipe github.com/khezen/espipe
echo "done!"

echo "purge..."
rm -rf /usr/local/go/
rm -rf /root/go/
rm -rf /tmp
apk del wget tar git ca-certificates
apk cache clean
echo "done!"

echo "instalation complete!"
