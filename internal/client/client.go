package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// Client handles WAF session management for chinadrugtrials.org.cn
type Client struct {
	httpClient *http.Client
	baseURL    string
	userAgent  string
}

// New creates a new Client with WAF session handling
func New() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		baseURL:   "https://www.chinadrugtrials.org.cn",
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36 Edg/150.0.0.0",
	}
}

// InitSession performs the initial GET to obtain WAF cookies.
func (c *Client) InitSession() error {
	req, err := http.NewRequest("GET", c.baseURL+"/clinicaltrials.searchlist.dhtml", nil)
	if err != nil {
		return fmt.Errorf("create session request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("init session request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// Wait for WAF challenge interval to allow cookies to become valid
	time.Sleep(3 * time.Second)
	return nil
}

// Get sends a GET request and handles WAF re-challenge.
func (c *Client) Get(path string) (int, []byte, error) {
	return c.getWithRetry(path, 1)
}

func (c *Client) getWithRetry(path string, retries int) (int, []byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("create get request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("get request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == 202 && retries > 0 {
		time.Sleep(3 * time.Second)
		if err := c.InitSession(); err != nil {
			return resp.StatusCode, body, fmt.Errorf("re-init session: %w", err)
		}
		return c.getWithRetry(path, retries-1)
	}

	return resp.StatusCode, body, nil
}

// Post sends a POST request with form data.
func (c *Client) Post(path string, params map[string]string) (int, []byte, error) {
	return c.postWithRetry(path, params, 2)
}

func (c *Client) postWithRetry(path string, params map[string]string, retries int) (int, []byte, error) {
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(formData.Encode()))
	if err != nil {
		return 0, nil, fmt.Errorf("create post request: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/clinicaltrials.searchlist.dhtml")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Microsoft Edge";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("post request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == 202 && retries > 0 {
		time.Sleep(3 * time.Second)
		if err := c.InitSession(); err != nil {
			return resp.StatusCode, body, fmt.Errorf("re-init session: %w", err)
		}
		return c.postWithRetry(path, params, retries-1)
	}

	return resp.StatusCode, body, nil
}

// PostDownload sends a POST for document download.
func (c *Client) PostDownload(path string, params map[string]string) (int, []byte, error) {
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(formData.Encode()))
	if err != nil {
		return 0, nil, fmt.Errorf("create download request: %w", err)
	}

	c.setDownloadHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("read download response: %w", err)
	}

	if resp.StatusCode == 202 {
		time.Sleep(3 * time.Second)
		if err := c.InitSession(); err != nil {
			return resp.StatusCode, body, fmt.Errorf("re-init session for download: %w", err)
		}
		return c.PostDownload(path, params)
	}

	return resp.StatusCode, body, nil
}

// SetCookie manually injects a cookie into the jar for pre-obtained session cookies.
func (c *Client) SetCookie(name, value string) {
	u, _ := url.Parse(c.baseURL)
	c.httpClient.Jar.SetCookies(u, []*http.Cookie{
		{Name: name, Value: value, Path: "/", Domain: ".chinadrugtrials.org.cn"},
	})
}

// CookieValue returns the value of a specific cookie by name.
func (c *Client) CookieValue(name string) string {
	u, _ := url.Parse(c.baseURL)
	for _, cookie := range c.httpClient.Jar.Cookies(u) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Microsoft Edge";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
}

func (c *Client) setDownloadHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("Origin", c.baseURL)
	req.Header.Set("Referer", c.baseURL+"/clinicaltrials.searchlistdetail.dhtml")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("sec-ch-ua", `"Not;A=Brand";v="8", "Chromium";v="150", "Microsoft Edge";v="150"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
}
