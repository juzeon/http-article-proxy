package article

import (
	"bytes"
	"encoding/base32"
	"http-article-proxy/data"
	"log"
	"strings"
)

func Encode(packets []data.Packet) (string, error) {
	var buf bytes.Buffer
	for _, packet := range packets {
		_, err := buf.WriteString(base32.StdEncoding.EncodeToString(packet.Data))
		if err != nil {
			return "", err
		}
		_, err = buf.WriteString("---")
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}
func Decode(content string) ([]data.Packet, error) {
	log.Println(content)
	if content == "" {
		return nil, nil
	}
	arr := strings.Split(content, "---")
	var packets []data.Packet
	for _, str := range arr[:len(arr)-1] {
		v, err := base32.StdEncoding.DecodeString(str)
		//log.Println("article decode: " + string(v))
		if err != nil {
			return nil, err
		}
		packets = append(packets, data.Packet{Data: v})
	}
	return packets, nil
}
