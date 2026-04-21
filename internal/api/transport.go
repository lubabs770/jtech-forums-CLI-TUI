package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type transport struct {
	baseURL    string
	httpClient *http.Client
	jar        *cookiejar.Jar
}

func newTransport(baseURL, sessionCookie string) (*transport, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	t := &transport{
		baseURL:    baseURL,
		jar:        jar,
		httpClient: &http.Client{Jar: jar},
	}
	if sessionCookie != "" {
		u, err := url.Parse(baseURL)
		if err != nil {
			return nil, err
		}
		jar.SetCookies(u, []*http.Cookie{{Name: "_t", Value: sessionCookie}})
	}
	return t, nil
}

func (t *transport) sessionCookie() string {
	u, _ := url.Parse(t.baseURL)
	for _, ck := range t.jar.Cookies(u) {
		if ck.Name == "_t" {
			return ck.Value
		}
	}
	return ""
}

func (t *transport) csrfToken() (string, error) {
	resp, err := t.httpClient.Get(t.baseURL + "/session/csrf.json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Csrf string `json:"csrf"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Csrf, nil
}

func (t *transport) login(username, password string) (string, error) {
	csrf, err := t.csrfToken()
	if err != nil {
		return "", fmt.Errorf("failed to get CSRF token: %w", err)
	}

	body, err := json.Marshal(map[string]string{"login": username, "password": password})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, t.baseURL+"/session", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if e.Error != "" {
			return "", fmt.Errorf("login failed: %s", e.Error)
		}
		return "", fmt.Errorf("login failed: status %d", resp.StatusCode)
	}

	for _, ck := range resp.Cookies() {
		if ck.Name == "_t" {
			return ck.Value, nil
		}
	}
	return "", fmt.Errorf("no session cookie in response")
}

func (t *transport) getJSON(path string, out any) error {
	resp, err := t.httpClient.Get(t.baseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &ErrUnauthorized{}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, path)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (t *transport) postJSON(path string, payload, out any) error {
	csrf, err := t.csrfToken()
	if err != nil {
		return fmt.Errorf("failed to get CSRF token: %w", err)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, t.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("X-CSRF-Token", csrf)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return &ErrUnauthorized{}
	}
	if resp.StatusCode >= 400 {
		var e struct {
			Errors []string `json:"errors"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		if len(e.Errors) > 0 {
			return fmt.Errorf("post failed: %s", e.Errors[0])
		}
		return fmt.Errorf("post failed: status %d", resp.StatusCode)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
