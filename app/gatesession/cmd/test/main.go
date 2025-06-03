package main

import (
	"fmt"
	"hash/fnv"
)

func usernameToID(username string) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte(username))
	return h.Sum64()
}

func main() {
	fmt.Println(usernameToID("tylyu"))
}
