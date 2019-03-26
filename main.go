package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/yuichitakeda/am-i-working/scape"
)

type loginInfo struct {
	User string
	Pass string
}

type empty struct{}

func saveToFile(fileName string, login loginInfo) error {
	data, encodeErr := json.MarshalIndent(login, "", "")
	if encodeErr != nil {
		return encodeErr
	}

	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readFile(fileName string) (loginInfo, error) {
	login := loginInfo{}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return login, err
	}
	decodeErr := json.Unmarshal(data, &login)
	if decodeErr != nil {
		return login, decodeErr
	}

	return login, nil
}

const configFile = "/home/yuichi/.scape_config.json"

func main() {

	pass := flag.String("p", "", "LDAP password")
	user := flag.String("u", "", "LDAP username")

	flag.Parse()

	login := loginInfo{User: *user, Pass: *pass}

	saveDone := make(chan empty)
	isLoginInfoEmpty := (login.User == "" || login.Pass == "")
	if isLoginInfoEmpty {
		loginInfoFromFile, err := readFile(configFile)
		if err != nil {
			fmt.Println("Must provide both user and password or use a valid config file")
			flag.Usage()
			return
		}
		login = loginInfoFromFile
	} else {
		go func() {
			err := saveToFile(configFile, login)
			if err != nil {
				fmt.Println("Error while saving to file")
			}
			saveDone <- empty{}
		}()
	}

	scape := scape.New()

	name := scape.Login(login.User, login.Pass)

	if name == "" {
		fmt.Println("Login failed")
		return
	}

	isWorking := scape.IsWorking(name)

	fmt.Println(isWorking)

	if !isLoginInfoEmpty {
		<-saveDone
	}
}
