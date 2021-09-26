# socket2tcp

> socket2tcp is a small tool that redirects local socket file traffic to the target tcp port; 
> This tool is used for development and debugging purposes, and its stability and performance
> are not reliable in a production environment.

## Install

Download the pre-compiled executable bin file or execute the `go install github.com/mritd/socket2tcp@latest` command.

## Usage

```sh
➜ socket2tcp --help
A simple tool for socket forwarding to remote tcp address

Usage:
  socket2tcp [flags]

Flags:
  -h, --help            help for socket2tcp
  -r, --remote string   remote tcp address
  -s, --socket string   unix socket address (default "/tmp/socket2tcp.sock")
```

## Example

```sh
# 1. Start a python simple http server(it listen on 0.0.0.0:8000 by default)
python3 -m http.server

# 2. Use socket2tcp to map the local socket file to port 8000 of the python3 simple http server
socket2tcp -s /tmp/test.sock -r 127.0.0.1:8000

# 3. Now use the curl command to access the local socket file, 
# socket2tcp will forward the traffic to the python3 simple http server
#
# Note: Any traffic sent to "/tmp/test.sock" will be forwarded to the tcp address of the "-r" option;
# The "http://127.0.0.1:8000" address at the end is just for curl to set the correct HTTP header,
# In actual use, we may not send HTTP traffic, it may be pure TCP traffic or other TCP-based protocols, 
# such as gRPC、FTP、SMTP, etc.
curl --no-buffer --unix-socket /tmp/test.sock http://127.0.0.1:8000
```
