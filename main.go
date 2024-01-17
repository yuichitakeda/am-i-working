package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/yuichitakeda/am-i-working/scape"
	"golang.org/x/term"
)

func If[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

func main() {
	u := flag.String("u", "", "LDAP username")

	flag.Parse()

	configFile := homeDir() + "/.scape_config.json"

	user := *u
	pass := ""
	if user != "" {
		fmt.Print("Password:")
		passbyte, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Print("Error reading password")
			return
		}

		pass = string(passbyte)
	}
	saveDone := make(chan struct{})
	isUserInfoEmpty := (user == "" || pass == "")
	if isUserInfoEmpty {
		usr, pss, err := RetrieveLoginInfo(configFile)
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

	workingDone := make(chan bool)
	go func() {
		isWorking := scapeApi.IsWorking(name)
		workingDone <- isWorking
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

	/*perc := float64(<-hoursMonthlyDone+time.Second) /
	float64(scape.GoalHours())
	*/
	//fmt.Println(<-workingDone, <-hoursDone, fmt.Sprintf("%.6f", perc*100))
	isWorking := <-workingDone

	fmt.Println("Working:", If(isWorking, "✅", "❌"))
	fmt.Println("Today:  ", <-hoursDone)
	fmt.Println("Monthly:", <-hoursMonthlyDone)
	if !isUserInfoEmpty {
		<-saveDone
	}
}
