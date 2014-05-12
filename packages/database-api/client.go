package main

import "time"
//import "fmt"

type Clerk struct {
  servers []string
  me int64
  seq int
}

type Connector struct {
  DNS string
}

type GetServerArgs struct {
}

type GetServerReply struct {
  Servers []string
  Err Err
}

func MakeClerk(servers []string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  ck.me = Nrand()
  ck.seq = 0
  return ck
}

func MakeConnector() *Connector {
  cn := new(Connector)
  cn.DNS = DNSaddress
  return cn
}

func (cn *Connector) GetServerList() []string {
  args := new(GetServerArgs)
  reply := new(GetServerReply)
  for {
    ok := call(cn.DNS, "DNSserver.GetServers", args, reply)
    if ok {
      return reply.Servers
    }
  }
}

// Gets the Pref List for the username and returns it.  Tries forever until successful
func (ck *Clerk) GetCoordinatorList(username string) []string {
  args := new(GetCoordListArgs)
  reply := new(GetCoordListReply)
  args.Username = username
  for {
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

// Does Coordinator Put and returns
func (ck *Clerk) CoordinatorPut(username string, message Message) {
  args := new(CoordPutArgs)
  reply := new(CoordPutReply)
  args.Username = username
  args.Data = message.Data
  args.Collection = message.Collection
  args.ID = message.Id
  args.IsHandoff = message.IsHandoff
  
  prefList := ck.GetCoordinatorList(username)
  
  for {
    for _, server := range prefList {
      ok := call(server, "MMDatabase.HandleCoordinatorPut", args, reply)
      if ok {
        if reply.Err == OK {
          return
        }
      }
      time.Sleep(50*time.Millisecond)
    }
    time.Sleep(500*time.Millisecond)
  }
}

func (ck *Clerk) Get(username string, id MessageID) Message {
  args := new(GetArgs)
  reply := new(GetReply)
  args.Username = username
  args.ID = id
  
  prefList := ck.GetCoordinatorList(username)
  
  for {
    for _, server := range prefList {
      ok := call(server, "MMDatabase.HandleGet", args, reply)
      if ok {
        if reply.Err == OK {
          return reply.Message
        }
      }
      time.Sleep(50*time.Millisecond)
    }
    time.Sleep(500*time.Millisecond)
  }
}
