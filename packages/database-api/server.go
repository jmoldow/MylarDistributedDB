package main

import "net"
import "fmt"
import "net/rpc"
import "log"
import "sync"
import "os"
import "syscall"
import "time"
import "hash/fnv"
import "math/big"
import "math/rand"
import cryptoRand "crypto/rand"
import "strconv"
import "encoding/gob"
import "encoding/json"

const (
  OK = "OK"
  ErrWrongCoordinator = "ErrWrongCoordinator"
  PUT = "PUT"
  GET = "GET"
  LIST = "LIST"
  DNSaddress = "/var/tmp/824-501/dns"
  InSocket = "/tmp/input.sock"
  Debug = 0
)

func DPrintf(format string, a ...interface{}) (n int, err error) {
  if Debug > 0 {
    log.Printf(format, a...)
  }
  return
}

/*
****************************************************
Data Types
****************************************************
*/

type MMDatabase struct {
  mu sync.Mutex
  l net.Listener
  me int
  dead bool // for testing
  unreliable bool // for testing
  servers []string
  nServers int
  nReplicas int // Number of replicas wanted
  handoffMessages []*Message // Messages that need to be handed off
  rpcCount int
}

type DNSserver struct {
  l net.Listener
  address string
  me string
  servers []string
}

type Message struct {
  Id MessageID
  // Whether or not this message needs to be handed off to another node later
  IsHandoff bool
  HandoffDestination string
  HandoffUsername string
  Data string
  Collection string
}

/*
****************************************************
API from Mylar/Meteor
****************************************************
*/

// Returns an ordered slice of servers in order they should be considered as coordinator
func (db *MMDatabase) GetCoordinatorList(username string) []string {
  initialIndex := db.getCoordinatorIndex(username)
  output := make([]string, 0)
  
  for i := initialIndex; i < len(db.servers); i++ {
    output = append(output, db.servers[i])
  }
  
  for i := 0; i < initialIndex; i++ {
    output = append(output, db.servers[i])
  }
  
  return output
}

// Returns success once nReplicas replicas are stored in the system
func (db *MMDatabase) CoordinatorPut(username string, message Message) Err {
//   Assert that this should be coordinator
//  if db.getCoordinatorIndex(username) != db.me && !message.IsHandoff {
//    return ErrWrongCoordinator
//  }
  
  totalReplicas := 0
  replicaLocations := make(map[int]bool)
  handoffTargets := make(map[int]bool)
  
  // Send to all N replicas except for this one (the coordinator)
  for totalReplicas < db.nReplicas-1 {
    for i, server := range(db.GetCoordinatorList(username)) {
      if !replicaLocations[i] && i != db.me {
        // Set Hinted Handoff
        handoffTarget := db.getHandoffTarget(username, i, replicaLocations, handoffTargets)
        if handoffTarget == -1 {
          message.IsHandoff = false
        } else {
          message.IsHandoff = true
          message.HandoffDestination = db.servers[i]
          message.HandoffUsername = username
          handoffTargets[handoffTarget] = true
        }
        // Set up args and reply
        args := new(ReplicaPutArgs)
        reply := new(ReplicaPutReply)
        args.Username = username
        args.Msg = message
        args.Handoff = false
        
        ok := call(server, "MMDatabase.ReplicaPut", args, reply)
      
        if ok && reply.Err == OK {
          totalReplicas++
          replicaLocations[i] = true
        }
      }
      
      if totalReplicas >=  db.nReplicas {
        break
      }
      
    }
  }
  
  // There should now be (at least) nReplicas-1 replicas in the system.
  // Replicate at the N-th server (this one / the coordinator),
  // then return success.
  // TODO: Handoff for coordinator if far down list
  db.LocalPut(username, message)
  totalReplicas++
  replicaLocations[db.me] = true
  
  return OK
}

func (db *MMDatabase) Get(username string, id MessageID) Message {
  // TODO: find message and return it
  return Message{}
}

/*
****************************************************
RPC Wrappers
****************************************************
*/

func (db *MMDatabase) HandleGetCoordinatorList(args *GetCoordListArgs, reply *GetCoordListReply) error {
  reply.PrefList = db.GetCoordinatorList(args.Username)
  reply.Err = OK
  return nil
}

func (db *MMDatabase) HandleCoordinatorPut(args *CoordPutArgs, reply *CoordPutReply) error {
  message := new(Message)
  message.Id = args.ID
  message.IsHandoff = args.IsHandoff
  message.Data = args.Data
  message.Collection = args.Collection
  
  reply.Err = db.CoordinatorPut(args.Username, *message)
  return nil
}

func (db *MMDatabase) HandleGet(args *GetArgs, reply *GetReply) error {
  message := db.Get(args.Username, args.ID)
  reply.Message = message
  reply.Err = OK
  return nil
}

/*
****************************************************
API to Mylar/Meteor
****************************************************
*/

func (db *MMDatabase) LocalPut(username string, msg Message) Err {
  // TODO
  return OK
}

func (db *MMDatabase) LocalDelete(username string, id MessageID) Err {
  // TODO
  return OK
}

/*
****************************************************
API to Servers
****************************************************
*/

func (db *MMDatabase) ReplicaPut(args *ReplicaPutArgs, reply *ReplicaPutReply) error {
  message := args.Msg
  // if message is satisfying a handoff, mark it as not needing handoff
  if args.Handoff {
    message.IsHandoff = false
  }
  
  // Do Local Put
  db.LocalPut(args.Username, message)
  
  // if message needs to be handed off, store in list of messages that need handing off
  if message.IsHandoff {
    db.handoffMessages = append(db.handoffMessages, &message)
  }
  reply.Err = OK
  return nil
}

func (dns *DNSserver) GetServers(args *GetServerArgs, reply *GetServerReply) error {
  reply.Servers = dns.servers
  reply.Err = OK
  return nil
}

/*
****************************************************
API Helpers
****************************************************
*/

// Returns a copy of slice without message at index
func removeMessage(slice []*Message, index int) []*Message {
  maxIndex := len(slice)-1
  
  lastElem := slice[maxIndex]
  slice[maxIndex] = slice[index]
  slice[index] = lastElem
  
  return slice[:maxIndex]
}

func (db *MMDatabase) runHandoffLoop() {
  for !db.dead {
    for i, message := range db.handoffMessages {
      // Set up args and reply
      args := new(ReplicaPutArgs)
      reply := new(ReplicaPutReply)
      args.Username = message.HandoffUsername
      args.Msg = *message
      args.Handoff = true
        
      ok := call(message.HandoffDestination, "MMDatabase.ReplicaPut", args, reply)
      
      if ok && reply.Err == OK {
        // Handoff successful, delete message
        db.handoffMessages = removeMessage(db.handoffMessages, i)
        break
      } else {
        time.Sleep(1000*time.Millisecond)
      }
    }
    time.Sleep(1000*time.Millisecond)
  }
}

// Returns index of first server that should be chosen as coordinator
func (db *MMDatabase) getCoordinatorIndex(username string) int {
  return int(hash(username) % uint32(db.nServers))
}

// Returns what the current handoff target should be with respect to replicaLocations
// Returns -1 if no handoff
// Assumes currentIndex is in range [0,nReplicas-1]
func (db *MMDatabase) getHandoffTarget(username string, currentIndex int, replicaLocations map[int]bool, handoffTargets map[int]bool) int {
  wrap := false
  firstReplica := db.getCoordinatorIndex(username)
  lastReplica := firstReplica + db.nReplicas
  if lastReplica >= db.nServers {
    wrap = true
    lastReplica = lastReplica % db.nServers
  }
  
  // Return -1 if in proper range
  if wrap {
    if currentIndex >= firstReplica || currentIndex <= lastReplica {
      return -1
    }
  } else {
    if firstReplica <= currentIndex && currentIndex <= lastReplica {
      return -1
    }
  }
  
  // Otherwise, target first one on priority list with no replica or targeted handoff yet
  i := firstReplica
  for {
    if !replicaLocations[i] && !handoffTargets[i] {
      return i
    }
    i++
    if i >= db.nServers {
      i = i % db.nServers
    }
  }
}

/*
****************************************************
Helper Functions
****************************************************
*/

func sameID(id1 RequestID, id2 RequestID) bool {
  return id1.ClientID == id2.ClientID && id1.Seq == id2.Seq
}

func hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}

func Nrand() int64 {
  max := big.NewInt(int64(1) << 62)
  bigx, _ := cryptoRand.Int(cryptoRand.Reader, max)
  x := bigx.Int64()
  return x
}

func port(host int) string {
  s := "/var/tmp/824/"
  os.Mkdir(s, 0777)
  s += "mmdb-"
  s += strconv.Itoa(host)
  return s
}

func portDNS() string {
  s := "/var/tmp/824/"
  os.Mkdir(s, 0777)
  s += "dns"
  return s
}

func cleanup(servers []*MMDatabase) {
  for i := 0; i < len(servers); i++ {
    if servers[i] != nil {
      servers[i].Kill()
    }
  }
}

//
// call() sends an RPC to the rpcname handler on server srv
// with arguments args, waits for the reply, and leaves the
// reply in reply. the reply argument should be a pointer
// to a reply structure.
//
// the return value is true if the server responded, and false
// if call() was not able to contact the server. in particular,
// the reply's contents are only valid if call() returned true.
//
// you should assume that call() will time out and return an
// error after a while if it doesn't get a reply from the server.
//
// please use call() to send all RPCs, in client.go and server.go.
// please don't change this function.
//
func call(srv string, name string, args interface{}, reply interface{}) bool {
  c, err := rpc.Dial("unix", srv)
  if err != nil {
    err1 := err.(*net.OpError)
    if err1.Err != syscall.ENOENT && err1.Err != syscall.ECONNREFUSED {
      fmt.Printf("paxos Dial() failed: %v\n", err1)
    }
    return false
  }
  defer c.Close()
    
  err = c.Call(name, args, reply)
  if err == nil {
    return true
  }

  fmt.Println(err)
  return false
}

/*
****************************************************
Helper Data Types
****************************************************
*/

type Err string

type MessageID int64

type ReplicaPutArgs struct {
  Username string
  Msg Message
  // Whether this ReplicaPut call is satisfying a Handoff (as opposed to being in top nReplicas of priority list)
  Handoff bool
}

type ReplicaPutReply struct {
  Err Err
}

type RequestID struct {
  ClientID int64
  Seq int64
}

type GetCoordListArgs struct {
  Username string
}

type GetCoordListReply struct {
  PrefList []string
  Err Err
}

type CoordPutArgs struct {
  ID MessageID
  IsHandoff bool
  Data string
  Collection string
  Username string
}

type CoordPutReply struct {
  Err Err
}

type GetArgs struct {
  Username string
  ID MessageID
}

type GetReply struct {
  Err Err
  Message Message
}

/*
****************************************************
Start and Kill Code
****************************************************
*/

// tell the server to shut itself down.
func (db *MMDatabase) Kill() {
  DPrintf("Kill(%d): die\n", db.me)
  db.dead = true
  db.l.Close()
}

//
// servers[] contains the ports of the set of
// servers that will cooperate via Paxos to
// form the fault-tolerant key/value service.
// me is the index of the current server in servers[].
// 
func StartServer(servers []string, me int, rpcs *rpc.Server) *MMDatabase {
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(Message{})
  gob.Register(GetCoordListArgs{})
  gob.Register(GetCoordListReply{})
  gob.Register(ReplicaPutArgs{})
  gob.Register(ReplicaPutReply{})
  gob.Register(CoordPutArgs{})
  gob.Register(CoordPutReply{})
  gob.Register(GetArgs{})
  gob.Register(GetReply{})

  db := new(MMDatabase)
  db.dead = false
  db.me = me
  db.servers = servers
  db.nServers = len(servers)
  db.nReplicas = 3
  db.handoffMessages = make([]*Message, 0)

  if rpcs != nil {
    // caller will create socket &c
    rpcs.Register(db)
  } else {
    rpcs = rpc.NewServer()
    rpcs.Register(db)

    os.Remove(servers[me])
    l, e := net.Listen("unix", servers[me]);
    if e != nil {
      log.Fatal("listen error: ", e);
    }
    db.l = l
    
    // please do not change any of the following code,
    // or do anything to subvert it.
    
    // create a thread to accept RPC connections
    go func() {
      for db.dead == false {
        conn, err := db.l.Accept()
        if err == nil && db.dead == false {
          if db.unreliable && (rand.Int63() % 1000) < 100 {
            // discard the request.
            conn.Close()
          } else if db.unreliable && (rand.Int63() % 1000) < 200 {
            // process the request but force discard of reply.
            c1 := conn.(*net.UnixConn)
            f, _ := c1.File()
            err := syscall.Shutdown(int(f.Fd()), syscall.SHUT_WR)
            if err != nil {
              fmt.Printf("shutdown: %v\n", err)
            }
            db.rpcCount++
            go rpcs.ServeConn(conn)
          } else {
            db.rpcCount++
            go rpcs.ServeConn(conn)
          }
        } else if err == nil {
          conn.Close()
        }
        if err != nil && db.dead == false {
          fmt.Printf("MMDatabase(%v) accept: %v\n", me, err.Error())
        }
      }
    }()
  }

  go db.runHandoffLoop()

  return db
}

func StartDNS(servers []string, rpcs *rpc.Server) *DNSserver {
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(DNSserver{})
  gob.Register(GetServerArgs{})
  gob.Register(GetServerReply{})

  dns := new(DNSserver)
  dns.servers = servers
  dns.address = portDNS()

  if rpcs != nil {
    // caller will create socket &c
    rpcs.Register(dns)
  } else {
    rpcs = rpc.NewServer()
    rpcs.Register(dns)

    os.Remove(dns.address)
    l, e := net.Listen("unix", dns.address);
    if e != nil {
      log.Fatal("listen error: ", e);
    }
    dns.l = l
    
    // please do not change any of the following code,
    // or do anything to subvert it.
    
    // create a thread to accept RPC connections
    go func() {
      for {
        conn, err := dns.l.Accept()
        if err == nil {
          go rpcs.ServeConn(conn)
        } 
        if err != nil {
          fmt.Printf("DNSerror: %v\n", err.Error())
        }
      }
    }()
  }

  return dns
}

// Starts up nservers servers and returns a slice of the objects and a slice of their ports
func RunServers(nservers int) ([]*MMDatabase, []string) {
  var kva []*MMDatabase = make([]*MMDatabase, nservers)
  var kvh []string = make([]string, nservers)

  for i := 0; i < nservers; i++ {
    kvh[i] = port(4000 + 3*i)
  }
  for i := 0; i < nservers; i++ {
    kva[i] = StartServer(kvh, i, nil)
  }
  
  return kva, kvh
}

/*
****************************************************
Main Function
****************************************************
*/

func main() {
  fmt.Println(len(os.Args))
  
  // Must specify client or server
  if len(os.Args) < 2 {
    fmt.Println("You must specify client or server")
    return
  }
  
  // Start Server
  if os.Args[1] == "server" {
    fmt.Println("Starting Servers")
    // If no extra argument given, default to nservers = 5
    nservers := 5
    // If argument given, start that many servers
    if len(os.Args) > 2 {
      lenArg := os.Args[2]
      num, err := strconv.Atoi(lenArg)
      if err != nil {
        fmt.Println("Non-Numeric nServers argument")
        return
      }
      nservers = num
    }
    
    // Start Servers
    servers, ports := RunServers(nservers)
    defer cleanup(servers)
    
    // Start DNS Server
    dns := StartDNS(ports, nil)
    fmt.Printf("DNS Info: %v\n", dns.servers)
    
    // Print ports and run until interrupted
    fmt.Printf("Server Ports: %v\n", ports)
    for {
      time.Sleep(1*time.Second)
    }
    
  // Start Client
  } else if os.Args[1] == "client" {
    fmt.Println("Getting Servers from DNS")
    cn := MakeConnector()
    serverList := cn.GetServerList()
    fmt.Println("Starting Client")
    ck := MakeClerk(serverList)
    
    // Connection Established, call methods on Clerk
//    prefList := ck.GetCoordinatorList("TestUser")
//    fmt.Printf("Pref List: %v\n", prefList)
    // TODO: issue set of commands
//    ck.CoordinatorPut("TestUser", Message{})
//    fmt.Println("Coordinator Put Finished")
//    ck.Get("TestUser", 0)
//    fmt.Println("Get Finished")

    for {
      ck.HandleRequest()
    }
  
  // Test
  } else if os.Args[1] == "test" {
    c, err := net.Dial("unix", InSocket)
    
    if err != nil {
      fmt.Println("error!\n")
    }
    
    requestString := `{"Type":"GET", "Username":"TestUser", "Collection":"MyCollection", "Data":"MyData"}`
    
    _, err = c.Write([]byte(requestString))  
    
    if err != nil {
      fmt.Println("error!\n")
    }
    
    buffer := make([]byte, 1024)
    var readlen int
    readlen, _ = c.Read(buffer)
    list := buffer[:readlen]
    answer := new(listreply)
    json.Unmarshal(list, answer)
    
    fmt.Printf("%v\n", answer.List)
  
  // Error
  } else {
    fmt.Println("Argument passed must be client or server")
    return
  }
}
