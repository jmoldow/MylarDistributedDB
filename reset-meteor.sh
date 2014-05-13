#!/bin/sh
meteor()
{
  xterm -e "cd ../MylarDistributedDB$1 && git fetch origin && git rebase origin/master && meteor reset" &
}

meteor 4000
meteor 4003
meteor 4006
meteor 4009
meteor 4012
