package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	dat, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	requestLine, err := parseRequestLine(dat)
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: requestLine,
	}, nil
}

func parseRequestLine(dat []byte) (RequestLine, error) {
	s := string(dat)
	result := strings.Split(s, "\r\n")[0]
	parts := strings.Fields(result)

	if len(parts) != 3 {
		return RequestLine{}, errors.New("invalid request line: wrong number of parts")
	}

	method := parts[0]
	target := parts[1]
	httpVersion := parts[2]

	if method != strings.ToUpper(method) || len(method) == 0 {
		return RequestLine{}, errors.New("invalid method")
	}

	if !strings.HasPrefix(httpVersion, "HTTP/") {
		return RequestLine{}, errors.New("invalid http version format")
	}

	version := strings.TrimPrefix(httpVersion, "HTTP/")
	if version != "1.1" {
		return RequestLine{}, errors.New("unsupported http version")
	}

	return RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}
