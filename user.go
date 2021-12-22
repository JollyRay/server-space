package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
)

type User struct {
	Id       uint32    `json:"id"`
	Roly     bool      `json:"roly"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Ban      bool      `json:"ban"`
	conn     *net.Conn `json:"-"`
}

type clientConteiner struct {
	List  []User     `json:"users"`
	mutex sync.Mutex `json:"-"`
}

func (us *clientConteiner) add(u User) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	for _, value := range us.List {
		if value.Login == u.Login {
			return false
		}
	}
	newId := us.List[len(us.List)-1].Id + 1
	u.Id = newId
	us.List = append(us.List, u)
	us.save()
	return true
}

func (us *clientConteiner) remove(id uint32) bool {
	fmt.Println("Remove user id ", id)
	us.mutex.Lock()
	defer us.mutex.Unlock()
	for index, value := range us.List {
		if value.Id == id {
			if value.conn == nil {
				us.List = append(us.List[:index], us.List[index+1:]...)
				return true
			}
			return false
		}
	}
	return false
}

func (us *clientConteiner) ban(id uint32, isBan bool) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	for index, value := range us.List {
		if value.Id == id {
			us.List[index].Ban = isBan
			if disconectAfterBan {
				tockens.banLogout(id)
			}
			return true
		}
	}
	return false
}

func (us *clientConteiner) getUser(login, password string) *User {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	if len(login) == 0 {
		return nil
	}
	for index, value := range us.List {
		if value.Login == login && value.Password == password {
			if value.Ban {
				fmt.Printf("User \"%s\" try login but is ban\n", login)
				return nil
			}
			if value.conn != nil {
				fmt.Printf("User \"%s\" login befer\n", login)
			}
			return &us.List[index]
		}
	}
	return nil
}

func (us *clientConteiner) save() {
	dataBytes, err := json.Marshal(us)
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile(jsonUserFilePath, dataBytes, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

var users clientConteiner

func initUser() {
	jsonFile, err := os.Open(jsonUserFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	jsonFile.Close()

	json.Unmarshal(byteValue, &users)
}
