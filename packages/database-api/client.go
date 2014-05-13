package main

import "time"
import "fmt"
import "net"
import "encoding/json"

type Clerk struct {
  servers []string
  me int64
  seq int
  port_in string
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

func MakeClerk(servers []string, port_in string) *Clerk {
  ck := new(Clerk)
  ck.servers = servers
  ck.me = Nrand()
  ck.seq = 0
  ck.port_in = port_in
  return ck
}

func RunClerks(servers []string) ([]*Clerk, []string) {
  nservers := len(servers)
  var kva []*Clerk = make([]*Clerk, nservers)
  var kvh []string = make([]string, nservers)

  for i := 0; i < nservers; i++ {
    kvh[i] = portIn(4000 + 3*i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = MakeClerk(servers, kvh[i])
    go func(i int) {
      for {
        kva[i].HandleRequest()
      }
    }(i)
  }
  return kva, kvh
}

func MakeConnector() *Connector {
  cn := new(Connector)
  cn.DNS = portDNS()
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

func (ck *Clerk) HandleRequest() {
  l, err := net.Listen("unix", ck.port_in)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  for {
    conn, err := l.Accept()
    if err != nil {
      fmt.Println(err)
    }
    
    buf := make([]byte, 4096)
    
    readlen, err := conn.Read(buf)
    
    request := string(buf[:readlen])
    
    fmt.Println(request)
    
    response := new(jsonreply)
    json.Unmarshal([]byte(request), response)
    
    fmt.Println(response)
    
    // Handle CoordinatorPut
    if response.Type == PUT {
      message := new(Message)
      message.Collection = response.Collection
      message.Data = response.Data
      id64 := int64(response.ID)
      var finalID MessageID = MessageID(id64)
      message.Id = finalID
      
      ck.CoordinatorPut(response.Username, *message)
      conn.Write([]byte(OK))
      continue
    }
    
    // Handle GetList
    output := ck.GetCoordinatorList(response.Username)
    listresponse := listreply{output}
    list, _ := json.Marshal(listresponse)
    
    conn.Write([]byte(list))
  }
}

type listreply struct {
  List []string
}

type jsonreply struct {
Type string
Username string
Collection string
Data string
ID int
}
