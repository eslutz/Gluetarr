package qbit

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	user    string
	pass    string
	client  *http.Client
}

type Preferences struct {
	ListenPort int `json:"listen_port"`
}

const (
	defaultHTTPTimeout   = 10 * time.Second
	requestRetryAttempts = 3
)

var requestRetryDelay = 2 * time.Second

func NewClient(baseURL, user, pass string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		user:    user,
		pass:    pass,
		client: &http.Client{
			Jar:     jar,
			Timeout: defaultHTTPTimeout,
		},
	}

	if err := client.Login(); err != nil {
		return nil, fmt.Errorf("initial login failed: %w", err)
	}

	return client, nil
}

func (c *Client) Login() error {
	data := url.Values{}
	data.Set("username", c.user)
	data.Set("password", c.pass)

	resp, err := c.client.PostForm(c.baseURL+"/api/v2/auth/login", data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer closeResponseBody(resp)

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(body) != "Ok." {
		return fmt.Errorf("login failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	slog.Debug("successfully authenticated with qBittorrent")
	return nil
}

func (c *Client) GetPort() (int, error) {
	var lastErr error
	for attempt := 1; attempt <= requestRetryAttempts; attempt++ {
		resp, err := c.doGet(c.baseURL + "/api/v2/app/preferences")
		if err != nil {
			lastErr = fmt.Errorf("failed to get preferences: %w", err)
		} else {
			port, decodeErr := decodePreferences(resp)
			if decodeErr == nil {
				return port, nil
			}
			lastErr = decodeErr
		}

		if attempt < requestRetryAttempts {
			slog.Warn("get port failed, retrying",
				"attempt", attempt,
				"max_attempts", requestRetryAttempts,
				"error", lastErr,
			)
			time.Sleep(requestRetryDelay)
		}
	}

	return 0, fmt.Errorf("failed to get preferences after %d attempts: %w", requestRetryAttempts, lastErr)
}

func (c *Client) SetPort(port int) error {
	prefsJSON := map[string]int{
		"listen_port": port,
	}

	jsonBytes, err := json.Marshal(prefsJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	data := url.Values{}
	data.Set("json", string(jsonBytes))

	var lastErr error
	for attempt := 1; attempt <= requestRetryAttempts; attempt++ {
		resp, err := c.doPostForm(c.baseURL+"/api/v2/app/setPreferences", data)
		if err != nil {
			lastErr = fmt.Errorf("failed to set preferences: %w", err)
		} else {
			body, readErr := io.ReadAll(resp.Body)
			closeResponseBody(resp)

			if resp.StatusCode == http.StatusOK {
				slog.Info("successfully updated qBittorrent listening port", "port", port)
				return nil
			}

			if readErr != nil {
				lastErr = fmt.Errorf("unexpected status code: %d, body read error: %w", resp.StatusCode, readErr)
			} else {
				lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			}
		}

		if attempt < requestRetryAttempts {
			slog.Warn("set port failed, retrying",
				"attempt", attempt,
				"max_attempts", requestRetryAttempts,
				"error", lastErr,
			)
			time.Sleep(requestRetryDelay)
		}
	}

	return fmt.Errorf("failed to set qBittorrent port after %d attempts: %w", requestRetryAttempts, lastErr)
}

func (c *Client) Ping() error {
	resp, err := c.doGet(c.baseURL + "/api/v2/app/version")
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) doGet(path string) (*http.Response, error) {
	resp, err := c.client.Get(path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusForbidden {
		return resp, nil
	}

	closeResponseBody(resp)
	slog.Warn("received 403 from qBittorrent, re-authenticating...")
	if err := c.Login(); err != nil {
		return nil, fmt.Errorf("re-authentication failed: %w", err)
	}

	return c.client.Get(path)
}

func (c *Client) doPostForm(path string, data url.Values) (*http.Response, error) {
	resp, err := c.client.PostForm(path, data)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusForbidden {
		return resp, nil
	}

	closeResponseBody(resp)
	slog.Warn("received 403 from qBittorrent, re-authenticating...")
	if err := c.Login(); err != nil {
		return nil, fmt.Errorf("re-authentication failed: %w", err)
	}

	return c.client.PostForm(path, data)
}

func decodePreferences(resp *http.Response) (int, error) {
	defer closeResponseBody(resp)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var prefs Preferences
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return 0, fmt.Errorf("failed to decode preferences: %w", err)
	}

	return prefs.ListenPort, nil
}

func closeResponseBody(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	if err := resp.Body.Close(); err != nil {
		slog.Warn("failed to close response body", "error", err)
	}
}
