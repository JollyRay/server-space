package main

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Linker struct {
	Car        *Car
	User       *User
	outsideCom int32
}

type tockenConteiner struct {
	List  map[int64]Linker
	mutex sync.Mutex
}

var tockens tockenConteiner = tockenConteiner{List: map[int64]Linker{}}

func (t *tockenConteiner) login(buf []byte, conn net.Conn) []byte {
	login, password := extractLoginPassword(buf)
	var user *User = users.getUser(login, password)
	if user == nil {
		return setNineByte(REFUSAL, 0)
	}
	user.conn = &conn
	var tocken int64
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for {
		tocken = rand.Int63()
		if _, ok := tockens.List[tocken]; !ok && tocken != 0 {
			break
		}
	}
	fmt.Printf("Login %s with password %s. Tocken generated %v or %d\n", login, password, convertToBytes(tocken), tocken)
	if user.Roly {
		tockens.List[tocken] = Linker{Car: nil, User: user}
		return setNineByte(ACCEPT, tocken)
	}
	var car *Car = cars.getFreeCar(tocken)
	var newLinker Linker = Linker{Car: car, User: user}
	tockens.List[tocken] = newLinker
	if car == nil {
		queue.take(tocken)
		return setNineByte(CONFIRMATION, tocken)
	}
	go imgRepeater(*user.conn, *car.conn, &car.stop, car.size, tocken)
	return setNineByte(ACCEPT, tocken)
}

func (t *tockenConteiner) logout(buf []byte) {
	var tocken int64 = convertToTocken(buf)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	linker, ok := t.List[tocken]
	if !ok {
		return
	}
	if linker.Car != nil {
		linker.Car.free(buf)
	}
	queue.leave(tocken)
	delete(t.List, tocken)

	if !autobancar {
		queue.next()
	}
}

func (t *tockenConteiner) banLogout(id uint32) {
	t.mutex.Lock()
	for tocken, linker := range t.List {
		if linker.User.Id == id {
			if linker.Car != nil {
				linker.Car.free(convertToBytes(tocken))
			}
			queue.leave(tocken)
			delete(t.List, tocken)
			if !autobancar {
				queue.next()
			}
			break
		}
	}
	t.mutex.Unlock()
}

func (t *tockenConteiner) sendData(buf []byte) {
	var tocken []byte = buf[1:9]
	linker, ok := t.List[convertToTocken(tocken)]
	if !ok {
		return
	}
	var x, y byte = buf[9], buf[10]
	request := append([]byte{SERVER_SEND}, tocken...)
	request = append(request, x, y)
	(*linker.Car.conn).Write(request)
}

func (t *tockenConteiner) freeCar(tocken int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	linker, ok := t.List[tocken]
	if !ok || linker.Car == nil {
		return
	}
	linker.Car.busy = false
	linker.Car.Ready = false
	linker.Car = nil
	t.List[tocken] = linker
}

func tockenDemon() {
	for {
		for _, value := range tockens.List {
			tockens.mutex.Lock()
			if value.outsideCom != 0 {
				if value.User.conn != nil {
					(*value.User.conn).SetReadDeadline(time.Now())
				}
			}
			tockens.mutex.Unlock()
			time.Sleep(time.Second)
		}
	}
}
