package lcu

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
)

const (
	lolProcessName = "LeagueClientUxRender.exe"
)

var (
	httpCli = &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	cli *client
)

type (
	client struct {
		port    int
		authPwd string
		baseUrl string
	}
	AssetResponse struct {
		StatusCode  int
		ContentType string
		Body        []byte
	}
)

func InitCli(port int, token string) {
	cli = NewClient(port, token)
}
func (cli client) fmtClientApiUrl() string {
	return fmt.Sprintf("https://riot:%s@127.0.0.1:%d", cli.authPwd, cli.port)
}
func NewClient(port int, token string) *client {
	cli := &client{
		port:    port,
		authPwd: token,
	}
	cli.baseUrl = cli.fmtClientApiUrl()
	return cli
}
func (cli client) httpGet(url string) ([]byte, error) {
	return cli.req(http.MethodGet, url, nil)
}
func (cli client) httpPost(url string, body interface{}) ([]byte, error) {
	return cli.req(http.MethodPost, url, body)
}
func (cli client) httpPatch(url string, body interface{}) ([]byte, error) {
	return cli.req(http.MethodPatch, url, body)
}
func (cli client) httpDel(url string) ([]byte, error) {
	return cli.req(http.MethodDelete, url, nil)
}
func (cli client) req(method string, url string, data interface{}) ([]byte, error) {
	var body io.Reader
	if data != nil {
		bts, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bts)
	}
	req, err := http.NewRequest(method, cli.baseUrl+url, body)
	if err != nil {
		return nil, err
	}
	if req.Body != nil {
		req.Header.Add("ContentType", "application/json")
	}
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (cli client) getWithIO(url string) (io.Reader, int64, error) {
	req, _ := http.NewRequest(http.MethodGet, cli.baseUrl+url, nil)
	if req.Body != nil {
		req.Header.Add("ContentType", "application/json")
	}
	resp, err := httpCli.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, resp.ContentLength, nil
}

func DetectAssetContentType(assetPath string, body []byte) string {
	if contentType := mime.TypeByExtension(filepath.Ext(assetPath)); contentType != "" {
		return contentType
	}
	if len(body) > 0 {
		return http.DetectContentType(body)
	}
	return "application/octet-stream"
}
