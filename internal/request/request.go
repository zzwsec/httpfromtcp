package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type RequestStat int

const (
	initialized RequestStat = iota
	done
)

const defaultBufferSize = 8

type Request struct {
	RequestLine RequestLine
	RequestStat RequestStat
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := &Request{RequestStat: initialized}

	// 初始化固定大小缓冲区
	buf := make([]byte, defaultBufferSize)
	readToIndex := 0

	for req.RequestStat != done {
		// 1. 缓冲区自动扩容 (如果已满)
		if readToIndex == len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		// 2. 从 Reader 读取数据，注意读取位置是从已存数据的末尾开始
		n, err := reader.Read(buf[readToIndex:])
		if n > 0 {
			readToIndex += n

			// 3. 尝试解析当前缓冲区中的所有有效数据
			consumed, parseErr := req.parse(buf[:readToIndex])
			if parseErr != nil {
				return nil, parseErr
			}

			// 4. 原地移动数据 (In-place buffer shift)
			// 如果解析了部分数据，将剩余未解析的数据移到最前面
			if consumed > 0 {
				copy(buf, buf[consumed:readToIndex])
				readToIndex -= consumed
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				// 只有在解析完成时 EOF 才是正常的
				if req.RequestStat != done {
					return nil, errors.New("incomplete request: connection closed")
				}
				break
			}
			return nil, err
		}
	}
	return req, nil
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
		// 已经 done 了就不该再传数据进来
		return 0, errors.New("error: trying to parse data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state %v", r.RequestStat)
	}
}

func parseRequestLine(dat []byte, rl *RequestLine) (int, error) {
	// 使用 bytes 库直接操作字节流，避免 string() 产生内存分配
	crlf := []byte("\r\n")
	idx := bytes.Index(dat, crlf)
	if idx == -1 {
		return 0, nil // 数据不完整，需要更多
	}

	line := dat[:idx]
	// 按照空白字符分割: [GET, /, HTTP/1.1]
	parts := bytes.Fields(line)
	if len(parts) != 3 {
		return 0, errors.New("invalid request line: wrong number of parts")
	}

	// 转换逻辑
	method := string(parts[0])
	target := string(parts[1])
	// versionRaw := string(parts[2])

	// 验证逻辑
	if !bytes.Equal(parts[0], bytes.ToUpper(parts[0])) {
		return 0, errors.New("invalid method")
	}
	if !bytes.HasPrefix(parts[1], []byte("/")) {
		return 0, errors.New("invalid target")
	}
	if !bytes.HasPrefix(parts[2], []byte("HTTP/")) {
		return 0, errors.New("invalid version format")
	}

	rl.Method = method
	rl.RequestTarget = target
	rl.HttpVersion = string(bytes.TrimPrefix(parts[2], []byte("HTTP/")))

	return idx + 2, nil // 消耗的长度 = 行长 + \r\n
}
