package article

import (
	"bytes"
	"encoding/base32"
)

func Encode(v []byte) (string, error) {
	var buf bytes.Buffer
	_, err := buf.WriteString(base32.StdEncoding.EncodeToString(v))
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
func Decode(content string) ([]byte, error) {
	if content == "" {
		return nil, nil
	}
	v, err := base32.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, err
	}
	return v, nil
}
