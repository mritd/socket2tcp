package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version, BuildDate, CommitID string

var socket, remote string
var version bool

var rootCmd = &cobra.Command{
	Use:     "socket2tcp",
	Short:   "A simple tool for socket forwarding to remote tcp address",
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		if version {
			showVersion()
		} else {
			run()
		}
	},
}

func run() {
	logrus.Infof("local socket forwarding to remote tcp: %s => %s", socket, remote)

	l, err := net.Listen("unix", socket)
	if err != nil {
		logrus.Fatal(err)
	}

	// Handle common process-killing signals, so we can gracefully shut down:
	ctx, stop := signal.NotifyContext(context.TODO(), os.Interrupt, os.Kill)
	defer stop()
	go func(ctx context.Context) {
		// Wait for a SIGINT or SIGKILL:
		<-ctx.Done()
		logrus.Info("Receive termination signal, gracefully shutdown.")
		// Stop listening (and unlink the socket if unix type):
		_ = l.Close()
	}(ctx)

	for {
		inConn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				logrus.Info("server shutdown.")
				return
			} else {
				logrus.Error(err)
				continue
			}
		}

		go func() {
			defer func() { _ = inConn.Close() }()
			outConn, err := net.Dial("tcp", remote)
			if err != nil {
				logrus.Error(err)
				return
			}
			defer func() { _ = outConn.Close() }()

			logrus.Infof("Handle conn %s => %s", socket, outConn.RemoteAddr())
			_, _, err = relay(inConn, outConn)
			if err != nil {
				logrus.Error(err)
			}
		}()
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initLog)
	rootCmd.Flags().StringVarP(&socket, "socket", "s", "/tmp/socket2tcp.sock", "unix socket address")
	rootCmd.Flags().StringVarP(&remote, "remote", "r", "", "remote tcp address")
	_ = rootCmd.MarkFlagRequired("socket")
	_ = rootCmd.MarkFlagRequired("remote")
	rootCmd.SetVersionTemplate(showVersion())
}

func initLog() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// relay copies between left and right bidirectionally. Returns number of
// bytes copied from right to left, from left to right, and any error occurred.
func relay(left, right net.Conn) (int64, int64, error) {
	type res struct {
		N   int64
		Err error
	}
	ch := make(chan res)

	go func() {
		n, err := io.Copy(right, left)
		_ = right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
		_ = left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
		ch <- res{n, err}
	}()

	n, err := io.Copy(left, right)
	_ = right.SetDeadline(time.Now()) // wake up the other goroutine blocking on right
	_ = left.SetDeadline(time.Now())  // wake up the other goroutine blocking on left
	rs := <-ch

	if err == nil {
		err = rs.Err
	}

	if e, ok := err.(net.Error); ok && e.Timeout() {
		err = nil // ignore i/o timeout
	}

	return n, rs.N, err
}

func showVersion() string {
	bannerBase64 := "ICAgICAgICAgICAgICAgIF8gICAgICAgIF8gICBfX19fXyAgXyAgICAgICAgICAgICAKICAgICAgICAgICAgICAgfCB8ICAgICAgfCB8IC8gX18gIFx8IHwgICAgICAgICAgICAKIF9fXyAgX19fICAgX19ffCB8IF9fX19ffCB8X2AnIC8gLyd8IHxfIF9fXyBfIF9fICAKLyBfX3wvIF8gXCAvIF9ffCB8LyAvIF8gXCBfX3wgLyAvICB8IF9fLyBfX3wgJ18gXCAKXF9fIFwgKF8pIHwgKF9ffCAgIDwgIF9fLyB8Xy4vIC9fX198IHx8IChfX3wgfF8pIHwKfF9fXy9cX19fLyBcX19ffF98XF9cX19ffFxfX1xfX19fXy8gXF9fXF9fX3wgLl9fLyAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIHwgfCAgICAKICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIHxffCAgICAK"
	versionTpl := `%s
Name: socket2tcp
Version: %s
Arch: %s
BuildDate: %s
CommitID: %s
`
	banner, _ := base64.StdEncoding.DecodeString(bannerBase64)
	return fmt.Sprintf(versionTpl, banner, Version, runtime.GOOS+"/"+runtime.GOARCH, BuildDate, CommitID)
}
