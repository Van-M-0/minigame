// +build darwin freebsd openbsd netbsd dragonfly
// +build !appengine

package mylog

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

type Termios syscall.Termios
