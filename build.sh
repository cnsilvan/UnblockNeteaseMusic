CurrentVersion=0.1.5
Project=github.com/cnsilvan/UnblockNeteaseMusic
Path="$Project/version"
ExecName="UnblockNeteaseMusic"
GitCommit=$(git rev-parse --short HEAD || echo unsupported)
GoVersion=$(go version)
BuildTime=$(date "+%Y-%m-%d %H:%M:%S")
platforms=("darwin/amd64")
buildGo() {
  GOOS=$1
  GOARCH=$2
  output_name=$ExecName
  if [ $GOOS = "windows" ]; then
    output_name+='.exe'
  fi
  echo "Building($GOOS/$GOARCH)..."
  TargetDir=bin/$GOOS/$GOARCH
  env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags "-X '$Path.Version=$CurrentVersion' -X '$Path.BuildTime=$BuildTime' -X '$Path.GoVersion=$GoVersion' -X '$Path.GitCommit=$GitCommit' -w -s" -o $TargetDir/$output_name
  if [ $? -ne 0 ]; then
    echo 'An error has occurred! Aborting the script execution...'
    exit 1
  fi
}
for platform in "${platforms[@]}"; do
  platform_split=(${platform//\// })
  buildGo ${platform_split[0]} ${platform_split[1]}
done
echo "--------------------------------------------"
echo "Version:" $CurrentVersion
echo "Git commit:" $GitCommit
echo "Go version:" $GoVersion
echo "Build Time:" $BuildTime
echo "Build Finish"
echo "--------------------------------------------"
