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
apk add wget tar git
echo "done!"

echo "building espipe..."
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
echo "done!"

echo "build complete!"
