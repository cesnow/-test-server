package sse

import (
    "fmt"
    "net/http"
    "strings"
    "time"
)

type Codec struct {
    headers map[string]string
}

func NewCodec() *Codec {
    return &Codec{
        headers: map[string]string{
            "Content-Type":                "text/event-stream",
            "Cache-Control":              "no-cache",
            "Connection":                 "keep-alive",
            "Access-Control-Allow-Origin": "*",
        },
    }
}

func (c *Codec) Encode(eventType string, data interface{}) ([]byte, error) {
    var builder strings.Builder
    
    if eventType != "" {
        builder.WriteString(fmt.Sprintf("event: %s\n", eventType))
    }
    
    switch v := data.(type) {
    case string:
        builder.WriteString(fmt.Sprintf("data: %s\n", v))
    case []byte:
        builder.WriteString(fmt.Sprintf("data: %s\n", string(v)))
    default:
        // 对于其他类型，可以使用 JSON 序列化
        builder.WriteString(fmt.Sprintf("data: %v\n", v))
    }
    
    builder.WriteString(fmt.Sprintf("id: %d\n", time.Now().UnixNano()))
    builder.WriteString("\n")
    
    return []byte(builder.String()), nil
}

func (c *Codec) WriteSSEHeaders(w http.ResponseWriter) {
    for key, value := range c.headers {
        w.Header().Set(key, value)
    }
}
