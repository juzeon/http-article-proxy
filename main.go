package main

import (
	"flag"
	"github.com/pkg/profile"
	_ "net/http/pprof"
)

func main() {
	typ := flag.String("type", "", "`server` or `client`")
	port := flag.Int("port", 0, "port to listen")
	url := flag.String("url", "", "url to request. Client only")
	dest := flag.String("dest", "", "destination to forward. Server only")
	flag.Parse()
	defer profile.Start(profile.CPUProfile,
		profile.ProfilePath("./profiles/"+*typ)).Stop()
	switch *typ {
	case "server":
		NewServer(*port, *dest).Serve()
	case "client":
		NewClient(*port, *url).Serve()
	default:
		panic("invalid type")
	}
}
