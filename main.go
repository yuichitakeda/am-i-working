package main

import (
	"amiworking/scape"
	"flag"
	"fmt"
)

var pass *string = flag.String("p", "", "LDAP password")
var user *string = flag.String("u", "", "LDAP username")

func main() {
	flag.Parse()

	u := *user
	p := *pass

	if u == "" || p == "" {
		fmt.Println("Must provide both user and password")
		flag.Usage()
		return
	}

	scape := scape.New()
	scape.Login(u, p)

	name := scape.Name()

	isWorking := scape.IsWorking(name)

	fmt.Println(isWorking)
}
