#!/BIN/BASH

DATE=$(date +"%Y-%m-%d|%H:%M:%S")
OS=$(uname| awk '{print tolower($0)}')
BUILD_INFO="[${DATE}]"
LDFLAGS="-X main.BuildInfo=${BUILD_INFO}"

APP=$1
echo "building $APP ..."
GOOS=$OS GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o output/"$APP"  main.go

cp -rf conf output/
cp run.sh output/