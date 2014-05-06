package mmdatabase

import "hash/fnv"

const (
  OK = "OK"
  ErrWrongCoordinator = "ErrWrongCoordinator"
)
type Err string

type Args struct {

}

type Reply struct {
  Err Err
}

type RequestID struct {
  ClientID int64
  Seq int64
}

type sameID(id1 RequestID, id2 RequestID) bool {
  return id1.ClientID == id2.ClientID && id1.Seq == id2.Seq
}

func hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}
