package main

import (
	"time"
)

var deley int64 = 100 // BETWEEN 0 - 1023

var timeout time.Duration = 10 * time.Minute

var autobanuser bool = true

var autobancar bool = true

var disconectAfterBan bool = false

var forceCarRemove bool = false

var jsonUserFilePath string = "basadata/users.json"

var jsonCarFilePath string = "basadata/car.json"

func setProperty(cartel []byte) {
	iter := 0
	for {
		if len(cartel) <= iter || cartel[iter]&0x80 == 0 {
			return
		}
		ruleNumber := (cartel[iter] & 0x7F) >> 2

		switch ruleNumber {
		case TIMEDELEY_NUM:
			if len(cartel) > iter+1 {
				deley = ((0x03 & int64(cartel[iter])) << 8) | int64(cartel[iter+1])
				iter++
			}
		case TIMEOUT_NUM:
			if len(cartel) > iter+2 {
				timeout = time.Duration((((rune(cartel[iter+1])) << 8) | rune(cartel[iter+2])) * 1e+9)
				iter++
				tockens.mutex.Lock()
				if timeout.Nanoseconds() == 0 {
					for _, value := range tockens.List {
						(*value.User.conn).SetReadDeadline(time.Time{})
					}
				} else {
				TIMEOUT_UPDATE:
					for key, value := range tockens.List {
						for _, tocken := range queue.Tockens {
							if key == tocken {
								continue TIMEOUT_UPDATE
							}
						}
						(*value.User.conn).SetReadDeadline(time.Now().Add(timeout))
					}
				}
				tockens.mutex.Unlock()
			}
		case CARBAN_NUM:
			if 0x03&cartel[iter] == 0 {
				autobancar = false
			} else {
				autobancar = true
			}
		case USERBAN_NUM:
			if 0x03&cartel[iter] == 0 {
				autobanuser = false
			} else {
				autobanuser = true
			}
		case FORCEDISCONNECTUSER_NUM:
			if 0x03&cartel[iter] == 0 {
				disconectAfterBan = false
			} else {
				disconectAfterBan = true
			}
		case FORCEDISCONNECTCAR_NUM:
			if 0x03&cartel[iter] == 0 {
				forceCarRemove = false
			} else {
				forceCarRemove = true
			}
		}
		iter++
	}
}
