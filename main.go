package main

import (
	"flag"
	"fmt"
	"os/user"
	"time"

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
		usr, pss, err := Retrieve(configFile)
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

	scapeApi := scape.New()

	name := scapeApi.Login(user, pass)

	if name == "" {
		fmt.Println("Login failed")
		return
	}

	workingDone := make(chan string)
	go func() {
		isWorking := scapeApi.IsWorking(name)
		workingDone <- fmt.Sprintf("%v", isWorking)
		close(workingDone)
	}()

	hoursDone := make(chan string)
	go func() {
		hours := scapeApi.HoursToday()
		hoursDone <- fmt.Sprintf("%v", hours)
		close(hoursDone)
	}()

	hoursMonthlyDone := make(chan time.Duration)
	go func() {
		hoursMonthly := scapeApi.HoursMonthly()
		hoursMonthlyDone <- hoursMonthly
		close(hoursMonthlyDone)
	}()

	perc := float64(<-hoursMonthlyDone+time.Second) / float64(scape.GoalHours())
	fmt.Println(<-workingDone, <-hoursDone, fmt.Sprintf("%.6f", perc*100))
	if !isLoginInfoEmpty {
		<-saveDone
	}
}
