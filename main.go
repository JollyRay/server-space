package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func handleConnection(conn net.Conn) {
	var karma rune = 0
	var mainTockenByte []byte = make([]byte, 8)
	var mainTocken int64
	var timeLast time.Time = time.Now()
	fmt.Println("Connection with", conn.RemoteAddr())
	defer conn.Close()
	reader := bufio.NewReader(conn)

CONNECTION:
	for {
		var request []byte = make([]byte, 1024)
		nread, err := reader.Read(request)

		if erprocess(err, &karma, mainTockenByte, timeLast) {
			tockens.logout(mainTockenByte)
			break CONNECTION
		}

		timeLast = time.Now()
		if _, ok := tockens.List[convertToTocken(request[1:9])]; request[0] != LOGIN && !ok {
			conn.Write(setNineByteAll(PERMISSIONDENIED, mainTockenByte))
			continue
		}

		switch request[0] {
		case LOGIN:
			respons := tockens.login(request[1:nread-1], conn)
			mainTockenByte = respons[1:]
			mainTocken = convertToTocken(mainTockenByte)
			replyNineRespons(respons, conn)
		case LOGOUT:
			if mainTocken == 0 || nread != 9 || comper(mainTockenByte, request[1:9]) {
				conn.Write(setNineByte(PERMISSIONDENIED, mainTocken))
				karma++
				continue CONNECTION
			}
			tockens.logout(request[1:])
		case DATA:
			if mainTocken == 0 || nread != 11 || comper(mainTockenByte, request[1:9]) {
				conn.Write(setNineByte(PERMISSIONDENIED, mainTocken))
				karma++
				continue CONNECTION
			}
			tockens.sendData(request[1:])
		case COMMAND:
			if mainTocken == 0 || !comper(mainTockenByte, request[1:9]) || !tockens.List[mainTocken].User.Roly {
				conn.Write(setNineByte(PERMISSIONDENIED, mainTocken))
				karma++
				continue CONNECTION
			}
			kind, info := extractCommand(request[9:nread])
			if executeManagerCommand(kind, info) {
				replyCommand(append([]byte{ACCEPT}, mainTockenByte...), kind, conn)
			} else {
				replyCommand(append([]byte{REFUSAL}, mainTockenByte...), kind, conn)
			}
		}
	}
}

func main() {
	initUser()
	initCar()
	rand.Seed(time.Now().Unix())

	PORT := ":3333"
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	go tockenDemon()
	go mimiManager()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go handleConnection(conn)
	}
}

func mimiManager() {
	reader := bufio.NewReader(os.Stdin)
	for {
		number, _ := reader.ReadString('\n')
		switch strings.TrimSpace(number) {
		case "1":
			fmt.Println("User:")
			fmt.Println(users.List)
		case "2":
			fmt.Println("Tocken:")
			fmt.Println(tockens.List)
		case "3":
			fmt.Println("Car:")
			fmt.Println(cars.List)
		case "4":
			fmt.Println("Queue:")
			fmt.Println(queue.Tockens)
		}
	}
}

/* Extract Function */

func extractLoginPassword(info []byte) (string, string) {
	var delimiter int
	for index, value := range info {
		if value == 0x0A {
			delimiter = index
			break
		}
	}
	if delimiter == 0 {
		return "", ""
	}
	return string(info[:delimiter]), strings.TrimSpace(string(info[delimiter+1:]))
}

func extractCommand(buf1 []byte) (kind byte, info interface{}) {
	kind = buf1[0]
	switch kind {
	case SHOWPILOTE, SHOWCAR, SHOWCLIENT, SHOWQUEUE, SHOWRULE:
		info = nil
	case ADDCLIENT, ADDMANAGER:
		var secondSpace int = 2
		for index, value := range buf1[1:] {
			if value == 0x0A {
				secondSpace = index
			}
		}
		for iter := secondSpace + 1; iter < len(buf1[1:]); iter++ {
			if buf1[iter] == 0x0A {
				secondSpace = iter
			}
		}
		login, password := extractLoginPassword(buf1[1:secondSpace])
		info = User{Login: login, Password: password, Roly: ADDMANAGER == kind, Ban: autobanuser}
	case ADDCAR:
		var elimiter int = 1
		for index, value := range buf1[1:] {
			if value == ':' {
				elimiter = index
			}
		}
		info = Car{Address: string(buf1[1:elimiter]), Port: string(buf1[elimiter+1:]), busy: false, Ready: !autobancar}
	case DELETECAR, DELETECLIENT, BANCAR, BANCLIENT, PREPARECAR, PREPARECLIENT:
		var index uint32 = 0
		for _, value := range buf1[1:5] {
			index <<= 8
			index |= uint32(value)
		}
		info = index
	case CLOSECONNECTION:
		info = convertToTocken(buf1[1:9])
	case SETRULE:
		info = buf1[1:]
	}

	return
}

/* Execute Manager Command*/

func executeManagerCommand(kind byte, info interface{}) bool {
	switch kind {
	case ADDCAR:
		car := info.(Car)
		return cars.add(car)
	case ADDCLIENT, ADDMANAGER:
		user := info.(User)
		return users.add(user)
	case DELETECAR:
		index := info.(uint32)
		return cars.remove(index)
	case DELETECLIENT:
		index := info.(uint32)
		return users.remove(index)
	case BANCAR:
		index := info.(uint32)
		return cars.ban(index, false)
	case BANCLIENT:
		index := info.(uint32)
		return users.ban(index, true)
	case PREPARECAR:
		index := info.(uint32)
		return cars.ban(index, true)
	case PREPARECLIENT:
		index := info.(uint32)
		return users.ban(index, false)
	case CLOSECONNECTION:
		tock := info.(int64)
		tockens.logout(convertToBytes(tock))
		return true
	case SETRULE:
		buf := info.([]byte)
		setProperty(buf)
		return true
	case SHOWPILOTE, SHOWCAR, SHOWCLIENT, SHOWQUEUE, SHOWRULE:
		return true
	}
	return false
}

/* Reply Function */

func replyNineRespons(buf []byte, conn net.Conn) {
	conn.Write(buf)
}

func replyCommand(answer []byte, kind byte, conn net.Conn) {
	if answer[0] == REFUSAL {
		conn.Write(answer)
	}
	switch kind {
	case ADDCAR, ADDCLIENT, ADDMANAGER, DELETECAR, DELETECLIENT, CLOSECONNECTION, BANCAR, BANCLIENT, PREPARECAR, PREPARECLIENT, SETRULE:
		conn.Write(answer)
	case SHOWCAR:
		var info string = ""
		for _, car := range cars.List {
			if car.busy {
				for _, linker := range tockens.List {
					if linker.Car != nil && linker.Car.Id == car.Id {
						info += fmt.Sprintf("%d,%s:%s,%t,%d;", car.Id, car.Address, car.Port, car.Ready, linker.User.Id)
					}
				}
			} else {
				info += fmt.Sprintf("%d,%s:%s,%t;", car.Id, car.Address, car.Port, car.Ready)
			}
		}
		answer = append(answer, []byte(info)...)
		answer = append(answer, byte(0))
		if value, ok := tockens.List[convertToTocken(answer[1:9])]; ok {
			fmt.Printf("Request's manager %s on SHOWCAR approved", value.User.Login)
		}
		conn.Write(answer)
	case SHOWCLIENT:
		var info string = ""
	LEAFLET:
		for _, user := range users.List {
			if user.conn == nil {
				info += fmt.Sprintf("%d,%s,%s,%t;", user.Id, user.Login, user.Password, user.Ban)
			} else {
				for tocken, linker := range tockens.List {
					if linker.User.Id == user.Id {
						if linker.Car != nil {
							info += fmt.Sprintf("%d,%s,%s,%t,%d,%d;", user.Id, user.Login, user.Password, user.Ban, linker.Car.Id, tocken)
						} else {
							info += fmt.Sprintf("%d,%s,%s,%t,%d;", user.Id, user.Login, user.Password, user.Ban, tocken)
						}
						continue LEAFLET
					}
				}
				info += fmt.Sprintf("%d,%s,%s,%t;", user.Id, user.Login, user.Password, user.Ban)
			}
		}
		answer = append(answer, []byte(info)...)
		answer = append(answer, byte(0))
		if value, ok := tockens.List[convertToTocken(answer[1:9])]; ok {
			fmt.Printf("Request's manager %s on SHOWUSER approved", value.User.Login)
		}
		conn.Write(answer)
	case SHOWPILOTE:
		var info string = ""
		for _, value := range tockens.List {
			if value.Car != nil {
				info += fmt.Sprintf("%d,%d;", value.User.Id, value.Car.Id)
			}
		}
		answer = append(answer, []byte(info)...)
		answer = append(answer, byte(0))
		if value, ok := tockens.List[convertToTocken(answer[1:9])]; ok {
			fmt.Printf("Request's manager %s on SHOWPILOTE approved", value.User.Login)
		}
		conn.Write(answer)
	case SHOWQUEUE:
		var info []byte
		for tockenConter := 1; tockenConter < len(queue.Tockens); tockenConter++ {
			info = append(info, convertToBytes(queue.Tockens[tockenConter])...)
		}
		answer = append(answer, info...)
		answer = append(answer, byte(0)) //TODO:  0 терм не самая хорошая идея
		if value, ok := tockens.List[convertToTocken(answer[1:9])]; ok {
			fmt.Printf("Request's manager %s on SHOWQUEUE approved", value.User.Login)
		}
		conn.Write(answer)
	case SHOWRULE:
		var info []byte = make([]byte, 3)
		if autobanuser {
			info[0] |= USERBAN
		}
		if autobancar {
			info[0] |= CARBAN
		}
		if disconectAfterBan {
			info[0] |= FORCEDISCONNECTUSER
		}
		if forceCarRemove {
			info[0] |= FORCEDISCONNECTCAR
		}
		info[0] |= byte(deley >> 6)
		info[1] |= byte(deley << 2)
		info[1] |= byte(int(timeout.Seconds()) >> 8)
		info[2] |= byte(timeout.Seconds())
		answer = append(answer, info...)
		if value, ok := tockens.List[convertToTocken(answer[1:9])]; ok {
			fmt.Printf("Request's manager %s on SHOWRULE approved", value.User.Login)
		}
		conn.Write(answer)
	}
}
