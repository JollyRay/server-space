package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"
)

func erprocess(err error, karma *rune, tocken []byte, last time.Time) bool {
	if err != nil || *karma > 9 {
		fmt.Println("Catch error with rule byte from", tocken, err.Error())
		*karma++
	}
	if time.Now().UnixMilli()-last.UnixMilli() < deley {
		fmt.Printf("Suspicious actions from tocken with deley %d millisec from %v", deley, tocken)
		*karma++
	}
	return checkEnd(err, *karma) || checkNeterr(err) || errors.Is(err, syscall.ECONNRESET)
}

func checkEnd(err error, karma rune) bool {
	if err == io.EOF || karma > 9 {
		fmt.Println("Count error strike", karma)
		return true
	}
	return false
}

func checkNeterr(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}

func imgHeaderCheck(buf []byte, err error, tocken int64) bool {
	if err != nil {
		fmt.Printf("Pipe breoken tocken: %d\n", tocken)
		return true
	}
	if len(buf) < 13 {
		fmt.Printf("Error header car. Messege so short (< 13 bytes). Tocken: %d", tocken)
		return true
	}
	if buf[0] != CAR_IMG {
		fmt.Printf("Error header car. Rule byte not right (!= CAR_IMG). Tocken: %d, %v", tocken, buf)
		return true
	}
	if !equals(buf[1:9], tocken) {
		fmt.Printf("Error header car. Tocken not equals. Tocken: %d, but received %d", tocken, convertToTocken(buf[1:9]))
		return true
	}
	if buf[9] == 0 && buf[10] == 0 && buf[11] == 0 {
		fmt.Printf("Error header car. Lenght img is zero. Tocken: %d", tocken)
		return true
	}
	return false
}
