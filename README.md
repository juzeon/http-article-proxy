# http-article-proxy

## Server

```bash
./http-article-proxy -type server -port 8080 -dest 127.0.0.1:10808
```
With an xray server running with the following configuration:
```json
{
    "logs": {
        "loglevel": "debug"
    },
    "inbounds": [
        {
            "listen": "127.0.0.1",
            "port": 10808,
            "protocol": "socks",
            "settings": {}
        }
    ],
    "outbounds": [
        {
            "protocol": "freedom",
            "tag": "direct"
        }
    ]
}
```

## Client

```bash
./http-article-proxy -type client -port 4444 -url http://server-ip:8080
```

Test:

```bash
curl https://cp.cloudflare.com -x socks5h://127.0.0.1:4444 -v
```

