# socket2tcp

## Install

`go install`

## Usage

```sh
➜ socket2tcp --help
A simple tool for socket forwarding to remote tcp addresses

Usage:
  socket2tcp [flags]

Flags:
  -h, --help            help for socket2tcp
  -r, --remote string   remote tcp address
  -s, --socket string   unix socket address (default "/tmp/socket2tcp.sock")
```

## Example

```sh
# start forwarding
socket2tcp -s /tmp/test.sock -r 127.0.0.1:8000

# start py simple http server
python3 -m http.server

# use curl to test it
curl --no-buffer --unix-socket /tmp/test.sock http://127.0.0.1:8080/
```