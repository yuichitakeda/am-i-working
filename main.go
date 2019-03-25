package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/yuichitakeda/am-i-working/scape"
)

type loginInfo struct {
	User string
	Pass string
}

var pass = flag.String("p", "", "LDAP password")
var user = flag.String("u", "", "LDAP username")

func saveToFile(fileName string, login loginInfo) error {
	data, err := json.MarshalIndent(login, "", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readFile(fileName string) (loginInfo, error) {
	file, err := os.Open(fileName)
	login := loginInfo{}
	if err != nil {
		return login, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decodeErr := decoder.Decode(&login)
	if decodeErr != nil {
		return login, decodeErr
	}

	return login, nil
}

const configFile = "/home/yuichi/.scape_config.json"

func main() {
	flag.Parse()

	login := loginInfo{User: *user, Pass: *pass}

	if login.User == "" || login.Pass == "" {
		loginInfo, err := readFile(configFile)
		if err != nil {
			fmt.Println("Must provide both user and password")
			flag.Usage()
			return
		}
		login = loginInfo
	}

	scape := scape.New()
	scape.Login(login.User, login.Pass)

	err := saveToFile(configFile, login)
	if err != nil {
		log.Fatal(err)
	}
	name := scape.Name()

	isWorking := scape.IsWorking(name)

	fmt.Println(isWorking)
}
