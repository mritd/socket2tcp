package main

import (
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var socket, remote string

var rootCmd = &cobra.Command{
	Use:   "socket2tcp",
	Short: "A simple tool for socket forwarding to remote tcp address",
	Run:   func(cmd *cobra.Command, args []string) { run() },
}

func run() {

	logrus.Infof("local socket forwarding to remote tcp: %s => %s", socket, remote)

	l, err := net.Listen("unix", socket)
	if err != nil {
		logrus.Fatal(err)
	}

	// Handle common process-killing signals so we can gracefully shut down:
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		// Wait for a SIGINT or SIGKILL:
		sig := <-c
		logrus.Infof("caught signal %s: shutting down.", sig)
		// Stop listening (and unlink the socket if unix type):
		err := l.Close()
		if err != nil {
			logrus.Error(err)
		}
	}(sigs)

	for {
		inConn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				logrus.Info("server shutdown.")
				os.Exit(0)
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
