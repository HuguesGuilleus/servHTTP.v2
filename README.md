# servHTTP.v2

My very simple web server.

# Installation

```bash
curl https://raw.githubusercontent.com/HuguesGuilleus/servHTTP.v2/master/install.bash | bash -s
```

# Configuration
File `/etc/servHTTP.ini` or the first args.

## General
```ini
; The log directory, each day a new file a created.
log = /var/log/servHTTP/
; disable the TLS listen and load certificates, use it when you don't have them.
notls = false
addr = :80
addrtls = :443
chalenge = /var/letsencrypt
```


## Certificate
```ini
[!cert.xxx]
key = ...
crt = ...
```

## Host
```ini
; File handler
[www.example.com]
; Two option for static file handler
/ = /var/www/
/ = file /var/www/
; you can define directory index template. I't Go HTML template that is executed
; on a slice of os.FileInfo
index = /path/to/template

; redirect handler
[example.com]
/ = redirect https://www.example.com

; Reverse proxy handler (work with web socket)
[reverse.example.com]
/ = reverse http://localhost:3000

; You can also define for each host, specific page for error
e404 = /path/to/error
e502 = /path/to/error
```
