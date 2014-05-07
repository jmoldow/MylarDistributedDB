package mmdatabase

import "net"
import "net/rpc"
import "log"
import "os"
import "syscall"
import "sync"
import "fmt"
import "math/rand"
import "strconv"
import "time"


type MMDatabase struct {
  mu sync.Mutex
  l net.Listener
  dead bool
  unreliable bool
  rpcCount int
  peers []string
  me int // index into peers[]

  // Your data here.
}

//
// tell the peer to shut itself down.
// for testing.
// please do not change this function.
//
func (db *MMDatabase) Kill() {
  db.dead = true
  if db.l != nil {
    db.l.Close()
  }
}

//
// the application wants to create a paxos peer.
// the ports of all the paxos peers (including this one)
// are in peers[]. this servers port is peers[me].
//
func Make(peers []string, me int, rpcs *rpc.Server) *MMDatabase {
  db := new(MMDatabase)
  db.peers = peers
  db.me = me

  if rpcs != nil {
    // caller will create socket &c
    rpcs.Register(db)
  } else {
    rpcs = rpc.NewServer()
    rpcs.Register(db)

    // prepare to receive connections from clients.
    // change "unix" to "tcp" to use over a network.
    os.Remove(peers[me]) // only needed for "unix"
    l, e := net.Listen("unix", peers[me]);
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


  return db
}
