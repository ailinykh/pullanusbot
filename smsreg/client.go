package smsreg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/logger"
)

const apiURL = "https://api.sms-reg.com"

// Client ...
type Client struct {
	Key    string
	client *http.Client
}

// NewClient ...
func NewClient(key string) *Client {
	client = &Client{Key: key, client: http.DefaultClient}
	return client
}

func (c *Client) getBalance() (*Balance, error) {
	data, err := c.Raw("getBalance")
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Balance
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) getList() (*List, error) {
	data, err := c.Raw("getList")
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp List
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) getNum(service string) (*Num, error) {
	data, err := c.Raw("getNum", fmt.Sprintf("service=%s", service))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Num
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) getState(tzid string) (*Tz, error) {
	data, err := c.Raw("getState", fmt.Sprintf("tzid=%s", tzid))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Tz
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) setReady(tzid string) (*Base, error) {
	data, err := c.Raw("setReady", fmt.Sprintf("tzid=%s", tzid))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Base
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) setOperationOkay(tzid string) (*Num, error) {
	data, err := c.Raw("setOperationOk", fmt.Sprintf("tzid=%s", tzid))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Num
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

func (c *Client) setOperationUsed(tzid string) (*Num, error) {
	data, err := c.Raw("setOperationUsed", fmt.Sprintf("tzid=%s", tzid))
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var resp Num
	if err = json.Unmarshal(data, &resp); err != nil {
		logger.Error(err)
		return nil, err
	}
	return &resp, nil
}

// Raw lets you call any method of sms API manually.
// It also handles API errors, so you only need to unwrap
// result field from json data.
func (c *Client) Raw(method string, params ...string) ([]byte, error) {

	url := apiURL + "/" + method + ".php?apikey=" + c.Key

	for _, p := range params {
		url = url + "&" + p
	}

	var buf bytes.Buffer
	// if err := json.NewEncoder(&buf).Encode(payload); err != nil {
	// 	return nil, err
	// }

	resp, err := c.client.Post(url, "application/json", &buf)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	resp.Close = true
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// returning data as well
	return data, nil
}
