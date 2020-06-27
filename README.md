# servHTTP.v2

My very simple web server.

# Installation

```bash
curl https://raw.githubusercontent.com/HuguesGuilleus/servHTTP.v2/master/install.bash | bash -s
```

# Configuration
File `/etc/servHTTP.ini`

## Certificate
```ini
[!cert.xxx]
key = ...
crt = ...
```

## Host
```ini
[example.com]
/ = redirect https://www.example.com

[www.example.com]
; Two option for file
/ = /var/www/
/ = file /var/www/

[reverse.example.com]
/ = reverse http://localhost:3000
```
