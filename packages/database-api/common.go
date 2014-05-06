package mmdatabase

import "hash/fnv"

const (
  OK = "OK"
)
type Err string

type Args struct {

}

type Reply struct {
  Err Err
}

func hash(s string) uint32 {
  h := fnv.New32a()
  h.Write([]byte(s))
  return h.Sum32()
}
