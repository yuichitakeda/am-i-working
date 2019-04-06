package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/zalando/go-keyring"
)


func storeUserInFile(user, filename string) error {
	jsonData, err := json.MarshalIndent(map[string]string{"user" : user}, "", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func storePassInKeyring(user, pass string) error {
	return keyring.Set("scape", user, pass)
}

func retrieveUserFromFile(fileName string) (string, error) {
	jsonData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	var jsonMap map[string]string
	err = json.Unmarshal(jsonData, &jsonMap)
	if err != nil {
		return "", err
	}

	return jsonMap["user"], nil
}

func retrievePassFromKeyring(user string) (string, error){
	return keyring.Get("scape", user)
}

func Store(filename, user, pass string) error {
	err := storeUserInFile(user, filename);
	if err != nil {
		return err
	}
	err = storePassInKeyring(user, pass)
	if(err != nil){
		return err
	}
	return nil
}

func Retrieve(filename string)(string, string, error){
	user, err := retrieveUserFromFile(filename);
	if err != nil {
		return "", "", err
	}

	pass, err := retrievePassFromKeyring(user)
	if err != nil {
		return "", "", err;
	}

	return user, pass, nil
}
