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


type Op struct {
}

type MMDatabase struct {
  mu sync.Mutex
  l net.Listener
  me int
  dead bool // for testing
  unreliable bool // for testing
}


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

