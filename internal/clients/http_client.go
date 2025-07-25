package clients

import (
	"net/http"
)

//go:generate mockgen -source=http_client.go -destination=mocks/mock_http_client.go -package=mocks

// HTTPClient is an interface that wraps the Do method of http.Client
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
