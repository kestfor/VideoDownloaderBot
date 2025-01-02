package cobalt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

var headers = http.Header{
	"Accept":       {"application/json"},
	"Content-Type": {"application/json"},
}

const (
	Tunnel   = "tunnel"
	Redirect = "redirect"
	Picker   = "picker"
	Error    = "error"
)

type CobaltInstance struct {
	apiUrl string
	client http.Client
}

type Response interface {
	GetStatus() string
	GetFileName() string
	GetUrl() string
}

func (t *TunnelResponse) GetStatus() string {
	return t.Status
}

func (t *TunnelResponse) GetFileName() string {
	return t.FileName
}

func (t *TunnelResponse) GetUrl() string {
	return t.Url
}

func (t *ErrorResponse) GetStatus() string {
	return t.Status
}

func (t *ErrorResponse) GetFileName() string {
	return ""
}

func (t *ErrorResponse) GetUrl() string {
	return ""
}

func (p *PickerResponse) GetStatus() string {
	return p.Status
}

func (p *PickerResponse) GetFileName() string {
	return p.AudioFilename
}

func (p *PickerResponse) GetUrl() string {
	return p.Picker[0].Url
}

type CobaltApi interface {
	FindVideo(url string) (Response, error)
	DownLoadVideo(searchRes Response) (*os.File, error)
}

func NewCobaltInstance(url string) *CobaltInstance {
	res := &CobaltInstance{apiUrl: url, client: http.Client{}}
	return res
}

func (c *CobaltInstance) FindVideo(url string) (Response, error) {
	body := map[string]string{
		"url": url,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("POST", c.apiUrl, bytes.NewBuffer(data))
	req.Header = headers
	response, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("api returned status %d error", response.StatusCode))
	}

	rawData := make([]byte, 4096)
	n, err := response.Body.Read(rawData)
	rawData = rawData[:n]
	mapped := make(map[string]string)

	err = json.Unmarshal(rawData, &mapped)
	if err != nil {
		return nil, err
	}

	status, ok := mapped["status"]
	if !ok {
		return nil, errors.New("no status key in json answer")
	}

	var res Response
	switch status {
	case Tunnel:
		res = &TunnelResponse{}
		err = json.Unmarshal(rawData, &res)
	case Redirect:
		res = &TunnelResponse{}
		err = json.Unmarshal(rawData, &res)
	case Picker:
		res = &PickerResponse{}
		err = json.Unmarshal(rawData, &res)
	case Error:
		res = &ErrorResponse{}
		err = json.Unmarshal(rawData, &res)
	}

	if res == nil {
		return nil, errors.New("there is no matched answer type")
	} else {
		return res, err
	}

}

func (c *CobaltInstance) setNilTransport() {
	c.client.Transport = nil
}

func (c *CobaltInstance) DownLoadVideo(searchRes Response) (*os.File, error) {
	var proxyUrl, _ = url.Parse(c.apiUrl)

	if searchRes.GetStatus() == Tunnel {
		c.client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		defer c.setNilTransport()
	}

	response, err := c.client.Get(searchRes.GetUrl())

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	buffer := make([]byte, 8096*64)
	file, err := os.Create(searchRes.GetFileName())
	if err != nil {
		return nil, err
	}

	n := 0
	var fileErr error = nil

	start := time.Now()
	timeOut := start.Add(time.Second * 10)

	for i := 0; err == nil && fileErr == nil; i++ {

		if i%10 == 0 && time.Now().After(timeOut) {
			file.Close()
			err = os.Remove(searchRes.GetFileName())
			return nil, errors.New("timeout while downloading file")
		}

		n, err = response.Body.Read(buffer)
		_, fileErr = file.Write(buffer[:n])
	}

	if err != io.EOF {
		file.Close()
		return nil, err
	}

	if fileErr != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}
