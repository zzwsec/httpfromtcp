package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}

	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}

	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	return n, nil
}

func TestRequestFromReader_Streaming(t *testing.T) {
	rawRequest := "GET /coffee HTTP/1.1\r\nHost: localhost\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"

	// 模拟从 1 字节到全长的所有分块读取可能性
	for i := 1; i <= len(rawRequest); i++ {
		t.Run("ReadSize_"+string(rune(i)), func(t *testing.T) {
			reader := &chunkReader{
				data:            rawRequest,
				numBytesPerRead: i,
			}

			r, err := RequestFromReader(reader)

			require.NoError(t, err, "Failed at chunk size: %d", i)
			require.NotNil(t, r)
			assert.Equal(t, "GET", r.RequestLine.Method)
			assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
		})
	}
}
