package mmdatabase

import "net"
import "fmt"
import "net/rpc"
import "log"
import "sync"
import "os"
import "syscall"
import "encoding/gob"
import "math/rand"
import "strconv"
import "time"

const Debug=0

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
}

type Message struct {
  isHandoff bool
  handoffDestination string
  timestamp = time.Time // Timestamp at which the coordinator first got the request
}

/*
****************************************************
API from Mylar/Meteor
****************************************************
*/

// Returns an ordered slice of peers in order they should be considered as coordinator
func (db *MMDatabase) GetCoordinatorList(username string) []string {
  initialIndex := getCoordinatorIndex(username)
  output := make([]string, 0)
  
  for i := initialIndex; i < len(db.peers); i++ {
    append(output, db.peers[i])
  }
  
  for i := 0; i < initialIndex; i++ {
    append(output, db.peers[i])
  }
  
  return output
}

// Returns success once nReplicas replicas are stored in the system
func (db *MMDatabase) CoordinatorPut(username string, id RequestID, message Message) Err {
  // Assert that this should be coordinator
  if db.getCoordinatorIndex(username) != db.me && !message.isHandoff {
    return ErrWrongCoordinator
  }
  
  message.timestamp = time.Now()
  totalReplicas := 0
  replicaLocations := make(map[int]bool)
  handoffTargets := make(map[int]bool)
  
  db.LocalPut(username, message)
  totalReplicas++
  replicaLocations[db.me] = true
  
  for totalReplicas < db.nReplicas {
    for i, server := db.GetCoordinatorList(username) {
      if !replicaLocations[i] {
        // Set Hinted Handoff
        handoffTarget := db.getHandoffTarget(username, i, replicaLocations, handoffTargets)
        if handoffTarget == -1 {
          // TODO: Set Handoff Nil
        } else {
          // TODO: Set Handoff to target
          handoffTargets[handoffTarget] = true
        }
        // TODO: Define args and reply
        ok := call(server, "MMDatabase.ReplicaPut", args, reply)
      
        if ok {
          if reply.Err == OK {
            totalReplicas++
            replicaLocations[i] = true
          }
        }
      }
      
      if totalReplicas >=  db.nReplicas {
        break
      }
      
    }
  }
  
  // There should now be (at least) nReplicas replicas in the system

  
  return OK
}

/*
****************************************************
API to Mylar/Meteor
****************************************************
*/

/*
****************************************************
API to Peers
****************************************************
*/

/*
****************************************************
API Helpers
****************************************************
*/

// Returns index of first peer that should be chosen as coordinator
func (db *MMDatabase) getCoordinatorIndex(username string) int {
  return hash(username) % db.nPeers
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
    db.lastReplica = db.lastReplica % db.nServers
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
Start and Kill Code
****************************************************
*/

// tell the server to shut itself down.
func (db *MMDatabase) kill() {
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
func StartServer(servers []string, me int) *MMDatabase {
  // call gob.Register on structures you want
  // Go's RPC library to marshall/unmarshall.
  gob.Register(Op{})

  db := new(MMDatabase)
  db.me = me
  db.servers = servers
  db.nServers = len(servers)
  db.nReplicas = 3

  // Your initialization code here.

  rpcs := rpc.NewServer()
  rpcs.Register(db)

  os.Remove(servers[me])
  l, e := net.Listen("unix", servers[me]);
  if e != nil {
    log.Fatal("listen error: ", e);
  }
  db.l = l


  // please do not change any of the following code,
  // or do anything to subvert it.

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
          go rpcs.ServeConn(conn)
        } else {
          go rpcs.ServeConn(conn)
        }
      } else if err == nil {
        conn.Close()
      }
      if err != nil && db.dead == false {
        fmt.Printf("MMDatabase(%v) accept: %v\n", me, err.Error())
        db.kill()
      }
    }
  }()

  return db
}