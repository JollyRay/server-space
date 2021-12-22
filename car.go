package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

type Car struct {
	Id      uint32    `json:"id"`
	Address string    `json:"address"`
	Port    string    `json:"port"`
	Ready   bool      `json:"ready"`
	busy    bool      `json:"-"`
	stop    bool      `json:"-"`
	size    int16     `json:"-"`
	conn    *net.Conn `json:"-"`
}

func (c *Car) setConnection(tocken int64) bool {
	var hello []byte = setNineByte(SERVER_HELLO, tocken)
	var answer []byte = make([]byte, 11)
	if c.conn != nil {
		_, err := (*c.conn).Write(hello)
		if err == nil {
			(*c.conn).SetDeadline(time.Now().Add(5 * time.Second))
			n, err := (*c.conn).Read(answer)
			fmt.Printf("Send car#%d message %v and get %v\n", c.Id, hello, answer)
			if err == nil && n == 11 && equals(answer[1:9], tocken) {
				(*c.conn).SetDeadline(time.Time{})
				var size uint16 = (uint16(answer[9]) << 8) | uint16(answer[10])
				c.size = int16(size)
				c.stop = false
				return true
			}
		}
		fmt.Printf("Car had broken connection %v\n", c)
	}
	conn, err := net.Dial("tcp", c.Address+":"+c.Port)
	if err == nil {
		c.conn = &conn
		(*c.conn).SetDeadline(time.Now().Add(5 * time.Second))
		(*c.conn).Write(hello)
		n, err := (*c.conn).Read(answer)
		fmt.Printf("Send car#%d message %v and get %v\n", c.Id, hello, answer)
		if err == nil && n == 11 && equals(answer[1:9], tocken) {
			(*c.conn).SetDeadline(time.Time{})
			var size uint16 = (uint16(answer[9]) << 8) | uint16(answer[10])
			c.size = int16(size)
			c.stop = false
			return true
		}
	}
	c.Ready = false
	return false
}

func (c *Car) free(tocken []byte) {
	if equals(tocken, 0) {
		return
	}
	c.busy = false
	c.Ready = false
	c.stop = true

	var finish []byte = append([]byte{SERVER_FINISH}, tocken...)
	if c.conn != nil {
		(*c.conn).Write(finish)
	}
}

type carsConteiner struct {
	List  []Car      `json:"models"`
	mutex sync.Mutex `json:"-"`
}

func (models *carsConteiner) add(c Car) bool {
	models.mutex.Lock()
	defer models.mutex.Unlock()
	for _, value := range models.List {
		if value.Address == c.Address && value.Port == c.Port {
			return false
		}
	}
	newId := models.List[len(models.List)-1].Id + 1
	c.Id = newId
	models.List = append(models.List, c)
	models.save()
	return true
}

func (models *carsConteiner) remove(id uint32) bool {
	models.mutex.Lock()
	defer models.mutex.Unlock()
	for index, value := range models.List {
		if value.Id == id {
			if !value.busy {
				if *value.conn != nil {
					(*value.conn).Close()
				}
				models.List = append(models.List[:index], models.List[index+1:]...)
				return true
			}
			return false
		}
	}
	return false
}

func (models *carsConteiner) ban(id uint32, isBan bool) bool {
	models.mutex.Lock()
	defer queue.next()
	defer models.mutex.Unlock()
	for index, value := range models.List {
		if value.Id == id {
			models.List[index].Ready = isBan
			if forceCarRemove {
				tockens.mutex.Lock()
				for key, value := range tockens.List {
					if value.Car.Id == id {
						queue.takeFirst(key)
						break
					}
				}
				tockens.mutex.Unlock()
			}
			return true
		}
	}
	return false
}

func (models *carsConteiner) getFreeCar(tocken int64) *Car {
	models.mutex.Lock()
	defer models.mutex.Unlock()

	for index, value := range models.List {
		if !value.busy && value.Ready && models.List[index].setConnection(tocken) {
			models.List[index].busy = true
			return &models.List[index]
		}
	}
	return nil
}

func (models *carsConteiner) save() {
	dataBytes, err := json.Marshal(models)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(jsonCarFilePath, dataBytes, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

var cars carsConteiner

func initCar() {
	jsonFile, err := os.Open(jsonCarFilePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close()
	json.Unmarshal(byteValue, &cars)
}

func imgRepeater(connU, connC net.Conn, stop *bool, size int16, tocken int64) {
	connC.Write(setNineByte(SERVER_START, tocken))
	var buf []byte = make([]byte, size)
	fmt.Println("Start read img from car size package", size)
REPEAT:
	for !*stop {
		nread, err := connC.Read(buf)
		if imgHeaderCheck(buf[:nread], err, tocken) {
			tockens.freeCar(tocken)
			queue.takeFirst(tocken)
			queue.next()
			break
		}
		buf[0] = CARIMG
		connU.Write(buf[:nread])
		var packSize int32 = (int32(buf[9])<<16 | int32(buf[10])<<8 | int32(buf[11])) - int32(nread) + carHeader
		fmt.Println(3, packSize+int32(nread)-carHeader, nread)
		// fmt.Println("Append:", buf[:nread])
		for packSize > 0 {
			nread, err := connC.Read(buf)
			if imgHeaderCheck(buf[:nread], err, tocken) {
				tockens.freeCar(tocken)
				queue.takeFirst(tocken)
				queue.next()
				break REPEAT
			}
			fmt.Println("Append:", nread, packSize)
			buf[0] = CARIMG
			connU.Write(buf[:nread])
			packSize -= int32(nread)
			// fmt.Println("Append:", buf[:nread])
		}
		fmt.Println("Finish")
	}
}
