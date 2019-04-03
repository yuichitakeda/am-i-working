package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os/user"

	"github.com/yuichitakeda/am-i-working/scape"
	"github.com/zalando/go-keyring"
)

type loginInfo struct {
	User string
	Pass string
}

type loginUser struct {
	User string
}

type empty struct{}

func saveToFile(fileName string, user string) error {
	data, encodeErr := json.MarshalIndent(loginUser{User: user}, "", "")
	if encodeErr != nil {
		return encodeErr
	}

	err := ioutil.WriteFile(fileName, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readFile(fileName string) (string, error) {
	login := loginUser{}

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	decodeErr := json.Unmarshal(data, &login)
	if decodeErr != nil {
		return "", decodeErr
	}

	return login.User, nil
}
func storeCredentials(login loginInfo) error {
	return keyring.Set("scape", login.User, login.Pass)
}

func retrieveCredentials(user string) (string, error) {
	return keyring.Get("scape", user)
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return usr.HomeDir
}

func loginInfoFail() {
	fmt.Println("Must provide both user and password or use a valid config file")
	flag.Usage()
}

func main() {
	p := flag.String("p", "", "LDAP password")
	u := flag.String("u", "", "LDAP username")

	flag.Parse()

	configFile := homeDir() + "/.scape_config.json"

	login := loginInfo{User: *u, Pass: *p}

	saveDone := make(chan empty)
	isLoginInfoEmpty := (login.User == "" || login.Pass == "")
	if isLoginInfoEmpty {
		loginUser, err := readFile(configFile)
		if err != nil {
			loginInfoFail()
			return
		}
		login.User = loginUser
		login.Pass, err = retrieveCredentials(login.User)
		if err != nil {
			loginInfoFail()
			return
		}
	} else {
		go func() {
			storeCredentials(login)
			saveToFile(configFile, login.User)
			saveDone <- empty{}
		}()
	}

	scape := scape.New()

	name := scape.Login(login.User, login.Pass)

	if name == "" {
		fmt.Println("Login failed")
		return
	}

	workingDone := make(chan string)
	go func() {
		isWorking := scape.IsWorking(name)
		workingDone <- fmt.Sprintf("%v", isWorking)
	}()

	hoursDone := make(chan string)
	go func() {
		hours := scape.HoursToday()
		hoursDone <- fmt.Sprintf("%v", hours)
	}()

	fmt.Println(<-workingDone, <-hoursDone)
	if !isLoginInfoEmpty {
		<-saveDone
	}
}
