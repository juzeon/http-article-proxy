# http-article-proxy

Server:

```bash
./http-article-proxy -type server -port 5555 -dest 127.0.0.1:1080 # a local socks5 proxy server uses the port 1080
```

Client:

```bash
./http-article-proxy -type client -port 4444 -url http://server-ip:5555
```

Test:

```bash
curl https://cp.cloudflare.com -x socks5h://127.0.0.1:4444 -v
```

