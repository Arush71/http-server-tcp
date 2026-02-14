package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewHeaders() Headers {
	return Headers{}
}

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("HOst"))
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("       Host: localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, 37, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing Headers
	headers = NewHeaders()
	data = []byte("  Host:  localhost:42069  \r\n   Day:  Monday  \r\n\r\n")
	read := 0
	isDone := false
	var isError error
	for isDone != true {
		n, done, err := headers.Parse(data[read:])
		isError = err
		if err != nil {
			break
		}
		read += n
		isDone = done
	}
	require.NoError(t, isError)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "Monday", headers.Get("DaY"))
	assert.Equal(t, 49, read)
	assert.True(t, isDone)

	// Test: Valid Done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	// assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: invalid field-name
	headers = NewHeaders()
	data = []byte(" H ost: lo ca lh ost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: invalid field-name wiht a special char
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	require.NotNil(t, headers)
	assert.NotEqual(t, "localhost:42069", headers.Get("H©st"))
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: multiple same field-name headers.
	headers = NewHeaders()
	headers.Set("Host", "Mac")
	data = []byte(" Host: Windows\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "Mac, Windows", headers.Get("Host"))
	assert.Equal(t, 14+2, n)
	assert.False(t, done)

}
