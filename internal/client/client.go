package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	// "net/http/httputil"
	"net/url"

	customerror "github.com/ahcogn/grafana-oncall-operator/internal/error"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

type Client interface {
	Get(context.Context, url.URL) ([]byte, error)
	Post(context.Context, url.URL, interface{}) ([]byte, error)
	Put(context.Context, url.URL, interface{}) ([]byte, error)
	Delete(context.Context, url.URL) error
}

type DefaultClient struct {
	Ctx context.Context
	Config
}

func (c *DefaultClient) Get(ctx context.Context, target url.URL) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.sendRequest(req)
	if err != nil {
		return nil, c.checkRespError(res, err)
	}

	raw, err := io.ReadAll(res.Body)
	return raw, c.checkRespError(res, err)

}

func (c *DefaultClient) Post(ctx context.Context, target url.URL, data interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(data)
	ctrllog.FromContext(ctx).Info(string(jsonStr))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, target.String(), bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, c.checkRespError(res, err)
	}

	raw, err := io.ReadAll(res.Body)
	ctrllog.FromContext(ctx).Info(string(raw))
	return raw, c.checkRespError(res, err)
}

func (c *DefaultClient) Put(ctx context.Context, target url.URL, data interface{}) ([]byte, error) {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, target.String(), bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return nil, c.checkRespError(res, err)
	}
	raw, err := io.ReadAll(res.Body)
	ctrllog.FromContext(ctx).Info(string(raw))
	return raw, c.checkRespError(res, err)
}

func (c *DefaultClient) Delete(ctx context.Context, target url.URL) error {
	req, err := http.NewRequest(http.MethodDelete, target.String(), nil)
	if err != nil {
		return err
	}
	res, err := c.sendRequest(req)
	return c.checkRespError(res, err)
}

func (c *DefaultClient) sendRequest(req *http.Request) (*http.Response, error) {
	token := c.Config.Credential

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	replyClient := &http.Client{}

	return replyClient.Do(req)
}

func (c *DefaultClient) checkRespError(resp *http.Response, err error) error {
	allowedResp := []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent}
	if err != nil {
		return customerror.HTTPError{
			HtErr: err,
		}
	}
	defer resp.Body.Close()
	// Check if resp allowed
	for _, code := range allowedResp {
		if code == resp.StatusCode {
			return nil
		}
	}

	return customerror.HTTPError{
		HtErr:      errors.New("invalid response"),
		StatusCode: resp.StatusCode,
	}
}
