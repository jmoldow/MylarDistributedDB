package main

import "time"
//import "fmt"

type Clerk struct {
  servers []string
  me int64
  seq int
}


func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  ck.me = Nrand()
  ck.seq = 0
  return ck
}

// Gets the Pref List for the username and returns it.  Tries forever until successful
func (ck *Clerk) GetCoordinatorList(username string) []string {
  for {
    args := new(GetCoordListArgs)
    reply := new(GetCoordListReply)
    args.Username = username
    for _, server := range ck.servers {
      ok := call(server, "MMDatabase.HandleGetCoordinatorList", args, reply)
      if ok && reply.Err == OK {
        return reply.PrefList
      }
      time.Sleep(50*time.Millisecond)
    }
    time.Sleep(500*time.Millisecond)
  }
}
