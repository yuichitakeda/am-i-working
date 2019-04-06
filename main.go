package main

import (
	"flag"
	"fmt"
	"os/user"

	"github.com/yuichitakeda/am-i-working/scape"
)

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func main() {
	p := flag.String("p", "", "LDAP password")
	u := flag.String("u", "", "LDAP username")

	flag.Parse()

	configFile := homeDir() + "/.scape_config.json"

	user, pass := *u, *p

	saveDone := make(chan struct{})
	isLoginInfoEmpty := (user == "" || pass == "")
	if isLoginInfoEmpty {
		usr, pss, err := Retrieve(configFile);
		if err != nil {
			fmt.Println("Must provide both user and password or use a valid config file and a keyring")
			flag.Usage()
			return
		}
		user, pass = usr, pss
	} else {
		go func() {
			Store(configFile, user, pass)
			saveDone <- struct{}{}
		}()
	}

	scape := scape.New()

	name := scape.Login(user, pass)

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
