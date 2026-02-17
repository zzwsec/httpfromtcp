package request

import (
	"errors"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	RequestStat RequestStat
}

type RequestStat int

const (
	initialized RequestStat = iota
	done
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{RequestStat: initialized}
	buffer := make([]byte, 0, 8) // 主缓冲区
	tmp := make([]byte, 8)       // 临时缓冲区

	for req.RequestStat != done {
		n, err := reader.Read(tmp)
		if n > 0 {
			// 1. 数据进入：追加到主缓冲区
			buffer = append(buffer, tmp[:n]...)

			// 2. 尝试解析：消耗主缓冲区的数据
			consumed, err := req.parse(buffer)
			if err != nil {
				return nil, err
			}

			// 3. 数据移出：删掉已经消耗的部分（进出移动）
			if consumed > 0 {
				buffer = buffer[consumed:]
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return req, nil
}

func parseRequestLine(dat []byte, rl *RequestLine) (int, error) {
	s := string(dat)
	exists := strings.Contains(s, "\r\n")
	if !exists {
		return 0, nil
	}

	result := strings.Split(s, "\r\n")[0]
	parts := strings.Fields(result)

	if len(parts) != 3 {
		return 0, errors.New("invalid request line: wrong number of parts")
	}

	method := parts[0]
	target := parts[1]
	httpVersion := parts[2]

	if method != strings.ToUpper(method) || len(method) == 0 {
		return 0, errors.New("invalid method")
	}

	if !strings.HasPrefix(target, "/") {
		return 0, errors.New("invalid http target format")
	}

	if !strings.HasPrefix(httpVersion, "HTTP/") {
		return 0, errors.New("invalid http version format")
	}

	version := strings.TrimPrefix(httpVersion, "HTTP/")
	if version != "1.1" {
		return 0, errors.New("unsupported http version")
	}
	rl.Method = method
	rl.RequestTarget = target
	rl.HttpVersion = version

	return len(result + "\r\n"), nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.RequestStat {
	case initialized:
		n, err := parseRequestLine(data, &r.RequestLine)
		if err != nil {
			return 0, err
		}
		if n > 0 {
			r.RequestStat = done
			return n, nil
		}
		return 0, nil
	case done:
		return 0, nil
	default:
		return 0, errors.New("unknown stat")
	}
}
