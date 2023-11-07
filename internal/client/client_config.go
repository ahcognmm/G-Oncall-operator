package client

import "net/url"

type Config struct {
	OncallUrl  url.URL
	Credential string
}
