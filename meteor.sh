#!/bin/sh
meteor()
{
  xterm -hold -e "cd ../MylarDistributedDB$1 && git fetch origin && git rebase origin/master && meteor --port=$1" &
}

go()
{
  xterm -hold -e "cd packages/database-api && go run server.go client.go $1" &
}

xterm -hold -e "cd ~/mylar/enc_modules/idp-email && meteor --port=3000" &
meteor 4000
meteor 4003
meteor 4006
meteor 4009
meteor 4012
go server
go client
