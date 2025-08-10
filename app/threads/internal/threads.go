package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL = "https://www.threads.net"
	apiURL  = "https://www.threads.net/api/graphql"

	getUserDocID        = "23996318473300828"
	getUserThreadsDocID = "9397645557011880"
	getUserRepliesDocID = "6307072669391286"
	getPostDocID        = "5587632691339264"
	getLikersDocID      = "9360915773983802"
)

var userIDRegex = regexp.MustCompile(`"user_id":"(\d+)"`)

type Client struct {
	client *http.Client
	header http.Header
	token  string
}

type Option func(*Client)

// NewClient returns a new threads API client.
// If no token is provided, it will fetch it automatically.
// If a token is provided, it will use it to make requests to the API.
// If a client is provided, it will be used to make requests to the API.
// If a header is provided, the original header will be modified to make requests to the API.
func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	c := Client{
		client: http.DefaultClient,
		header: make(http.Header),
	}
	c.header.Add("Authority", "www.threads.net")
	c.header.Add("Accept", "*/*")
	c.header.Add("Accept-Language", "en-US,en;q=0.9")
	c.header.Add("Cache-Control", "no-cache")
	c.header.Add("Content-Type", "application/x-www-form-urlencoded")
	c.header.Add("Connetion", "keep-alive")
	c.header.Add("Origin", "https://www.threads.net")
	c.header.Add("Pragma", "no-cache")
	c.header.Add("Sec-Fetch-Site", "same-origin")
	c.header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.44.639.844 Safari/537.36")
	c.header.Add("X-IG-ASBD-ID", "129477")
	c.header.Add("X-IG-App-ID", "238260118697367")

	for _, opt := range opts {
		opt(&c)
	}

	if c.token == "" {
		token, err := c.getToken(ctx)
		if err != nil {
			return nil, err
		}
		c.token = token
	}
	c.header.Add("X-FB-LSD", c.token)

	return &c, nil
}

// WithToken returns an option that sets the token used to make requests to the API.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithClient returns an option that sets the client used to make requests to the API.
func WithClient(client *http.Client) Option {
	return func(c *Client) {
		c.client = client
	}
}

// WithHeader returns an option that sets the header used to make requests to the API.
func WithHeader(header http.Header) Option {
	return func(c *Client) {
		c.header = header
	}
}

// getToken returns the token used to make requests to the API.
func (c *Client) getToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/@instagram", nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.44.639.844 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	pos := bytes.Index(body, []byte("\"token\""))
	return string(body[pos+9 : pos+31]), nil
}

// GetUserID returns the user ID of the given username.
func (c *Client) GetUserID(ctx context.Context, name string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/@"+name, nil)
	if err != nil {
		return 0, err
	}
	req.Header = c.header.Clone()
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Referer", baseURL)
	req.Header.Add("Sec-Fetch-Dest", "document")
	req.Header.Add("Sec-Fetch-Mode", "navigate")
	req.Header.Add("Sec-Fetch-Site", "cross-site")
	req.Header.Add("Sec-Fetch-User", "?1")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	req.Header.Del("X-Asbd-Id")
	req.Header.Del("X-Fb-Lsd")
	req.Header.Del("X-Ig-App-Id")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(userIDRegex.FindSubmatch(body)[1]))
}

// GetUser returns the profile of the given user ID.
func (c *Client) GetUser(ctx context.Context, userID int) ([]byte, error) {
	h := c.header.Clone()
	h.Add("X-FB-Friendly-Name", "BarcelonaProfileRootQuery")
	return sendRequest(ctx, c.client, h, c.token,
		getUserDocID,
		map[string]int{"userID": userID},
	)
}

// GetUserThreads returns the threads posted by the given user ID.
func (c *Client) GetUserThreads(ctx context.Context, userID int) ([]byte, error) {
	h := c.header.Clone()
	h.Add("X-FB-Friendly-Name", "BarcelonaProfileThreadsTabQuery")
	return sendRequest(ctx, c.client, h, c.token,
		getUserThreadsDocID,
		map[string]int{"userID": userID},
	)
}

// GetUserReplies returns the replies posted by the given user ID.
func (c *Client) GetUserReplies(ctx context.Context, userID int) ([]byte, error) {
	h := c.header.Clone()
	h.Add("X-FB-Friendly-Name", "BarcelonaProfileRepliesTabQuery")
	return sendRequest(ctx, c.client, h, c.token,
		getUserRepliesDocID,
		map[string]int{"userID": userID},
	)
}

// GetPost returns the post of the given post ID.
func (c *Client) GetPost(ctx context.Context, postID int) ([]byte, error) {
	h := c.header.Clone()
	h.Add("X-FB-Friendly-Name", "BarcelonaPostPageQuery")
	return sendRequest(ctx, c.client, h, c.token,
		getPostDocID,
		map[string]int{"postID": postID},
	)
}

// GetLikers returns the liker list of the given post ID.
func (c *Client) GetLikers(ctx context.Context, postID int) ([]byte, error) {
	return sendRequest(ctx, c.client, c.header, c.token,
		getLikersDocID,
		map[string]int{"mediaID": postID},
	)
}

func sendRequest(ctx context.Context, c *http.Client, headers http.Header, token, docID string, variables map[string]int) ([]byte, error) {
	b, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}

	data := url.Values{}
	data.Set("lsd", token)
	data.Set("doc_id", docID)
	data.Set("variables", string(b))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = headers

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// ---- GraphQL 呼叫 ----

func (c *Client) sendRequest(ctx context.Context, docID string, variables map[string]any) ([]byte, error) {
	b, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Set("lsd", c.token)
	data.Set("doc_id", docID)
	data.Set("variables", string(b))
	// 可選：帶 jazoest（與 lsd 搭配提高成功率）
	data.Set("jazoest", jazoest(c.token))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	h := c.header.Clone()
	// 友好名稱（可有可無；方便伺服端識別 query）
	h.Set("X-FB-Friendly-Name", "BarcelonaProfileThreadsTabQuery")
	req.Header = h

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func jazoest(lsd string) string { // 常見做法：runecode 累加，前綴 '2'
	sum := 0
	for _, r := range lsd {
		sum += int(r)
	}
	return "2" + strconv.Itoa(sum)
}

// GetAllUserThreads 會連續呼叫 GetUserThreadsPage，使用回應中的 end_cursor 逐頁抓到沒有下一頁
// first 建議 10~50；maxRetries 針對暫時性錯誤；delay 每頁間隔(限速)
func (c *Client) GetAllUserThreads(ctx context.Context, userID int, first, maxRetries int, delay time.Duration) ([][]byte, error) {
	if first <= 0 {
		first = 15
	}
	if maxRetries < 0 {
		maxRetries = 0
	}
	if delay < 0 {
		delay = 0
	}

	var pages [][]byte
	var after string
	page := 1

	for {
		raw, err := c.callWithRetry(ctx, maxRetries, func() ([]byte, error) {
			return c.GetUserThreadsPage(ctx, userID, first, after)
		})
		if err != nil {
			return pages, fmt.Errorf("GetUserThreads page=%d: %w", page, err)
		}
		pages = append(pages, raw)

		next, hasNext, err := parseCursor(raw)
		if err != nil {
			// doc_id 若不支援分頁或回應格式變動，這裡會回錯；先把已抓頁面回給你
			return pages, fmt.Errorf("parseCursor page=%d: %w", page, err)
		}
		if !hasNext || next == "" {
			break
		}
		after = next
		page++

		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return pages, ctx.Err()
			}
		}
	}
	return pages, nil
}

// 盡量兼容不同回應形狀的 paging 定位
func parseCursor(raw []byte) (endCursor string, hasNext bool, err error) {
	type paging struct {
		EndCursor   string `json:"end_cursor"`
		HasNextPage bool   `json:"has_next_page"`
	}

	// A: 頂層 paging
	var a struct {
		Paging *paging `json:"paging"`
	}
	if json.Unmarshal(raw, &a) == nil && a.Paging != nil {
		return a.Paging.EndCursor, a.Paging.HasNextPage, nil
	}

	// B: data.paging
	var b struct {
		Data struct {
			Paging *paging `json:"paging"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &b) == nil && b.Data.Paging != nil {
		return b.Data.Paging.EndCursor, b.Data.Paging.HasNextPage, nil
	}

	// C: data.mediaData.paging
	var c struct {
		Data struct {
			MediaData struct {
				Paging *paging `json:"paging"`
			} `json:"mediaData"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &c) == nil && c.Data.MediaData.Paging != nil {
		return c.Data.MediaData.Paging.EndCursor, c.Data.MediaData.Paging.HasNextPage, nil
	}

	// D: 廣義掃描（有些層級不同）
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err == nil {
		if ec, hn, ok := scanPagingMap(m); ok {
			return ec, hn, nil
		}
	}
	return "", false, fmt.Errorf("paging not found")
}

func scanPagingMap(m map[string]any) (string, bool, bool) {
	// 嘗試在當前層找 paging
	if p, ok := m["paging"].(map[string]any); ok {
		ec, _ := p["end_cursor"].(string)
		hn, _ := p["has_next_page"].(bool)
		if ec != "" {
			return ec, hn, true
		}
	}
	// 淺遞迴尋找
	for _, v := range m {
		if mm, ok := v.(map[string]any); ok {
			if ec, hn, ok := scanPagingMap(mm); ok {
				return ec, hn, true
			}
		}
	}
	return "", false, false
}

func (c *Client) callWithRetry(ctx context.Context, maxRetries int, fn func() ([]byte, error)) ([]byte, error) {
	backoff := 400 * time.Millisecond
	for i := 0; i <= maxRetries; i++ {
		raw, err := fn()
		if err == nil {
			return raw, nil
		}
		es := strings.ToLower(err.Error())
		retryable := strings.Contains(es, "429") || strings.Contains(es, "rate") ||
			strings.Contains(es, "timeout") || (strings.Contains(es, "http") && strings.Contains(es, "5"))
		if i < maxRetries && retryable {
			select {
			case <-time.After(backoff):
				backoff *= 2
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, err
	}
	return nil, fmt.Errorf("unreachable")
}

// GetUserThreadsPage 允許你指定每頁筆數(first)與翻頁游標(after)
func (c *Client) GetUserThreadsPage(ctx context.Context, userID int, first int, after string) ([]byte, error) {
	h := c.header.Clone()
	h.Add("X-FB-Friendly-Name", "BarcelonaProfileThreadsTabQuery")

	vars := map[string]any{
		"userID": userID,
	}
	if first > 0 {
		vars["first"] = first
	}
	if after != "" {
		vars["after"] = after
	}
	return c.sendRequest(ctx, getUserThreadsDocID, vars)
}

// 回傳 true 代表這個 doc_id 接受 after/first，false 代表不吃（要換 doc_id）
func (c *Client) DocIDSupportsAfter(ctx context.Context, docID string, userID int) bool {
	vars := map[string]any{
		"userID": userID,
		"first":  2,
		"after":  "DUMMY_CURSOR",
	}
	body, err := c.sendRequest(ctx, docID, vars)
	if err != nil {
		// 有些會回 200 + error JSON；我們掃字串就好
		es := strings.ToLower(err.Error())
		fmt.Println(es)
		// 若明確回「Unknown argument 'after'」或 schema error，就不支援
		if strings.Contains(es, "unknown argument") || strings.Contains(es, "schema") {
			return false
		}
		// 其他錯誤先視為不支援
		return false
	}

	fmt.Println(string(body))
	// 如果 200，看看回應裡是否有 "end_cursor" 這類字眼
	return bytes.Contains(body, []byte("end_cursor")) || bytes.Contains(body, []byte("endCursor")) ||
		bytes.Contains(body, []byte(`"paging"`)) || bytes.Contains(body, []byte(`"page_info"`))
}
