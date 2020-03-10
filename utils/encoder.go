package utils

import (
	"regexp"
	"fmt"
	"encoding/hex"
	"log"
	"encoding/base64"
	"strings"
)

func getParams(regEx, text string) (paramsMap map[string]string) {

	var compRegEx = regexp.MustCompile(regEx)
	match := compRegEx.FindStringSubmatch(text)

	paramsMap = make(map[string]string)
	for i, name := range compRegEx.SubexpNames() {
		fmt.Print(name)
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	return paramsMap
}

func EncodeUserId(userId string) (string) {
	params := getParams(`U(?P<id>[0-9a-f]{32})`, userId)
	id := params["id"]

	decoded, err := hex.DecodeString(id)
	if err != nil {
		log.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(decoded))
	s := strings.Split(encoded, "=")[0]
	s = strings.Replace(s, "+", "-", -1)
	s = strings.Replace(s, "/", "_", -1)
	fmt.Printf("%s\n", s)

	return s
}

func DecodeUserId(encode string) (string) {
	s := encode
	s = strings.Replace(s, "-", "+", -1)
	s = strings.Replace(s, "_", "/", -1)
	switch len(s) % 4 {
		case 0: break
		case 2: s += "=="; break
		case 3: s += "="; break
	}
	decode, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}
	id := hex.EncodeToString(decode)
	userId := fmt.Sprintf("U%s", id)
	return userId
}

func ExtractEncodeUserId(text string) (string) {
	params := getParams(`\((?P<id>[A-Za-z0-9+_]+)\)`, text)
	encode := params["id"]
	log.Printf("encode: %s", encode)
	userId := DecodeUserId(encode)
	return userId
}