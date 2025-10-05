//go:build !windows

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

func openTTY() (*os.File, func(), error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return os.Stdin, func() {}, nil
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0)
	if err != nil {
		return nil, func() {}, fmt.Errorf("no TTY available")
	}
	return tty, func() { _ = tty.Close() }, nil
}

func makeRaw(f *os.File) (restore func(), err error) {
	st, err := term.MakeRaw(int(f.Fd()))
	if err != nil {
		return nil, err
	}
	return func() { _ = term.Restore(int(f.Fd()), st) }, nil
}

func characterChan(r *os.File) <-chan rune {
	out := make(chan rune, 16)
	go func() {
		defer close(out)
		br := bufio.NewReader(r)
		for {
			ch, _, err := br.ReadRune()
			if err != nil {
				return
			}
			out <- ch
		}
	}()
	return out
}

func ListenForQuit(parent context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)

		// ctx that cancels on SIGINT/SIGTERM or parent cancel
		ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
		defer stop()

		tty, closeTTY, err := openTTY()
		if err != nil {
			<-ctx.Done()
			return
		}
		defer closeTTY()

		restore, err := makeRaw(tty)
		if err != nil {
			return
		}
		defer func() { restore(); fmt.Print("\r\n") }()

		keys := characterChan(tty)

		for {
			select {
			case <-ctx.Done():
				return
			case ch, ok := <-keys:
				if !ok {
					return
				}
				switch ch {
				case 27, 'q', 'Q', rune(3): // ESC, q/Q, Ctrl-C
					return
				default:
					if ch >= 32 && ch < 127 {
						fmt.Printf("\rpressed: %q\r\n", ch)
					} else {
						fmt.Printf("\rpressed: 0x%02X\r\n", byte(ch))
					}
				}
			}
		}
	}()
	return done
}
