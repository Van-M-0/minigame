
export -p GOOS=windows 
export -p GOARCH=amd64
export -p GOPATH=$PWD:/Users/van/go
export -p GOBIN=$PWD/bin

#set GO_ENABLED=0 
#set GOOS=windows 
#set GOARCH=amd64

#set GOPATH=/Users/van/work/minigame:/Users/van/go

#go build -x -v game
/usr/local/go/bin/go build -x -i -o -v -o /Users/van/work/minigame/bin/server.exe app
