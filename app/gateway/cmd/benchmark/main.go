package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"google.golang.org/protobuf/proto"
	"kiyudesign.com/cesnow/light-server/pkg/atomic"
	"kiyudesign.com/cesnow/light-server/pkg/tproto"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	url         = "ws://localhost:7824" // WebSocket server URL
	numMessages = 1                     // Number of messages per client
)

// LatencyStats keeps track of latency data
type LatencyStats struct {
	totalLatency time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
	totalCount   int
	successCount int
	failCount    int
	mu           sync.Mutex
}

// AddLatency records latency and success/failure stats
func (ls *LatencyStats) AddLatency(latency time.Duration, success bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	ls.totalLatency += latency
	ls.totalCount++

	// Track min and max latency
	if ls.minLatency == 0 || latency < ls.minLatency {
		ls.minLatency = latency
	}
	if latency > ls.maxLatency {
		ls.maxLatency = latency
	}

	// Track success/fail counts
	if success {
		ls.successCount++
	} else {
		ls.failCount++
	}
}

var msgIdSeq = atomic.NewAtomicInt64(0)

func nextMessageId() int64 {
	unixNano := time.Now().UnixNano()
	ts := unixNano / 1e9
	ms := (unixNano % 1e9) / 1e6
	sid := msgIdSeq.Add(1) & 0x1ffff
	msgIdSeq.CompareAndSwap(0x1ffff, 0)
	msgId := ts<<32 | int64(ms)<<21 | sid<<3 | int64(1)
	return msgId
}

// AverageLatency calculates average latency
func (ls *LatencyStats) AverageLatency() time.Duration {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	if ls.successCount == 0 {
		return 0
	}
	return ls.totalLatency / time.Duration(ls.successCount)
}

func makeTestPayload() []byte {
	initDataBytes := make([]byte, 32)
	nextMsgId := nextMessageId()
	binary.LittleEndian.PutUint64(initDataBytes[0:], uint64(0))             // auth_id [1, 8]
	binary.LittleEndian.PutUint64(initDataBytes[8:], uint64(0))             // session_id [2, 8]
	binary.LittleEndian.PutUint32(initDataBytes[8+8:], uint32(0))           // seq_no [3, 4]
	binary.LittleEndian.PutUint64(initDataBytes[8+8+4:], uint64(nextMsgId)) // msg_id [4, 8]

	message := &tproto.InitConnection{}
	checksum := tproto.GetTProtoChecksum(message)

	messageBytes, _ := proto.Marshal(message)
	checksumBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(checksumBytes[0:], uint32(checksum)) // checksum [5, 4]
	payloadBytes := append(checksumBytes, messageBytes...)             // payload [6, ...]

	binary.LittleEndian.PutUint32(initDataBytes[8+8+8+4:], uint32(len(payloadBytes))) // payload length
	sendDataBytes := append(initDataBytes, payloadBytes...)
	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes[0:], uint32(len(sendDataBytes))) // prefix with all size [0, 4]
	return append(sizeBytes, sendDataBytes...)
}

var initConnectionData []byte

func init() {
	peek := 0x39F57B94
	peekData := make([]byte, 4)
	binary.BigEndian.PutUint32(peekData, uint32(peek))
	initConnectionData = append(peekData, makeTestPayload()...)
}

// testWebSocket simulates a WebSocket connection
func testWebSocket(stats *LatencyStats, wg *sync.WaitGroup) {
	defer wg.Done()

	// Establish WebSocket connection
	conn, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		log.Println("Connection failed:", err)
		stats.AddLatency(0, false)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "testing complete")

	sendData := initConnectionData
	for i := 0; i < numMessages; i++ {
		start := time.Now()

		if i > 0 {
			sendData = makeTestPayload()
		}
		// Send message
		// log.Printf("%s [%d]\n", hex.EncodeToString(sendData), len(sendData))
		err := conn.Write(context.Background(), websocket.MessageBinary, sendData)
		if err != nil {
			log.Println("Write error:", err)
			stats.AddLatency(0, false)
			return
		}

		// Receive message
		_, _, err = conn.Read(context.Background())

		latency := time.Since(start)

		if err != nil {
			log.Println("Read error:", err)
			stats.AddLatency(0, false)
		} else {
			stats.AddLatency(latency, true)
		}
	}
}

func main() {
	// Use flag to set the number of clients
	numClients := flag.Int("clients", 5, "Number of concurrent WebSocket clients")
	flag.Parse()

	var wg sync.WaitGroup
	stats := &LatencyStats{}

	// Start performance test
	startTime := time.Now()
	for i := 0; i < *numClients; i++ {
		wg.Add(1)
		go testWebSocket(stats, &wg)
	}
	wg.Wait()

	// Calculate and display results
	totalDuration := time.Since(startTime)
	averageLatency := stats.AverageLatency()
	throughput := float64(stats.successCount) / totalDuration.Seconds()

	fmt.Printf("Number Message By Client: %v\n", numMessages)
	fmt.Printf("Total Duration: %v\n", totalDuration)
	fmt.Printf("Average Latency: %v\n", averageLatency)
	fmt.Printf("Minimum Latency: %v\n", stats.minLatency)
	fmt.Printf("Maximum Latency: %v\n", stats.maxLatency)
	fmt.Printf("Total Messages Sent: %d\n", stats.totalCount)
	fmt.Printf("Successful Messages: %d\n", stats.successCount)
	fmt.Printf("Failed Messages: %d\n", stats.failCount)
	fmt.Printf("Throughput (Messages per Second): %.2f\n", throughput)
}
