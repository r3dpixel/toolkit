package reqx

import (
	"io"

	"github.com/imroc/req/v3"
	"github.com/r3dpixel/toolkit/stringsx"
)

// Bytes extracts bytes from response - use for chaining: client.R().Get(url).Bytes()
func Bytes(resp *req.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return resp.ToBytes()
}

// String extracts string from response - use for chaining: client.R().Get(url).String()
func String(resp *req.Response, err error) (string, error) {
	bytes, err := Bytes(resp, err)
	if err != nil {
		return "", err
	}

	return stringsx.FromBytes(bytes), nil
}

// Stream extracts stream from response
func Stream(resp *req.Response, err error) (io.ReadCloser, error) {
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
