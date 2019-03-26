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
func storeCredentials(login loginInfo) {
	keyring.Set("scape", login.User, login.Pass)
}

func retrieveCredentials(user string) string {
	p, err := keyring.Get("scape", user)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return p
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return usr.HomeDir
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
			fmt.Println("Must provide both user and password or use a valid config file")
			flag.Usage()
			return
		}
		login.User = loginUser
		login.Pass = retrieveCredentials(login.User)
	} else {
		go func() {
			storeCredentials(login)
			err := saveToFile(configFile, login.User)
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
