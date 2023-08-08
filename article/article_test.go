package article

import (
	"http-article-proxy/data"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		packets []data.Packet
		want    string
		wantErr bool
	}{
		{
			name:    "empty slice",
			packets: []data.Packet{},
			want:    "",
			wantErr: false,
		},
		{
			name: "single packet",
			packets: []data.Packet{
				{Data: []byte("Hello")},
			},
			want:    "JBSWY3DP---",
			wantErr: false,
		},
		{
			name: "multiple packets",
			packets: []data.Packet{
				{Data: []byte("Hello")},
				{Data: []byte("World")},
				{Data: []byte("!")},
			},
			want:    "JBSWY3DP---K5XXE3DE---EE======---",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.packets)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []data.Packet
		wantErr bool
	}{
		{
			name:    "empty string",
			content: "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "single packet",
			content: "JBSWY3DP---",
			want: []data.Packet{
				{Data: []byte("Hello")},
			},
			wantErr: false,
		},
		{
			name:    "multiple packets",
			content: "JBSWY3DP---K5XXE3DE---EE======---",
			want: []data.Packet{
				{Data: []byte("Hello")},
				{Data: []byte("World")},
				{Data: []byte("!")},
			},
			wantErr: false,
		},
		{
			name:    "invalid content",
			content: "JBSWY3DP---JBSWY3DPEBLW64TMMQ---IQ---XYZ---",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decode(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Decode() = %v, want %v", got, tt.want)
			}
		})
	}
}
