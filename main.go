package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func getUser() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

func hash(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

var (
	User                 = getUser()
	StoreFilePath string = filepath.Join(
		User.HomeDir, ".vault-tokens",
	)
	VAULT_ADDR string = os.Getenv("VAULT_ADDR")
	hashedAddr string = hash(VAULT_ADDR)
	gitRef     string = ""
	gitTag     string = ""
)

type TokenStore struct {
	FilePath string                 `yaml:"file_path"`
	Data     map[string]interface{} `yaml:"data"`
}

func (ts *TokenStore) Load() {
	if _, err := os.Stat(ts.FilePath); err == nil {
		content, err := ioutil.ReadFile(ts.FilePath)
		if err != nil {
			panic(err)
		}
		yaml.Unmarshal(content, &ts.Data)
	} else {
		ts.Data = map[string]interface{}{}
	}

}

func usage() {
	fmt.Fprintf(os.Stderr,
		`%s %s-%s
	
	important: this token helper is not meant to be executed directly
	supported commands: get, store, erase
`, os.Args[0], gitTag, gitRef,
	)
}

func get(ts *TokenStore) {
	Token, _ := ts.Data[hashedAddr].(string)
	fmt.Fprintf(os.Stdout, "%s", Token)
}
func store(ts *TokenStore) {
	reader := bufio.NewReader(os.Stdin)
	token, _ := reader.ReadString('\n')
	ts.Data[hashedAddr] = token
}

func erase(ts *TokenStore) {
	_, ok := ts.Data[hashedAddr]
	if ok {
		delete(ts.Data, hashedAddr)
	}
}

func sync(ts *TokenStore) {
	dataBytes, err := yaml.Marshal(ts.Data)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(StoreFilePath, dataBytes, 0400)
}

func main() {
	if len(VAULT_ADDR) == 0 {
		fmt.Fprintln(os.Stderr, "err: VAULT_ADDR is unset")
		os.Exit(100)
	}
	if len(os.Args) >= 2 {
		ts := &TokenStore{FilePath: StoreFilePath}
		ts.Load()
		switch os.Args[1] {
		case "get":
			get(ts)
		case "store":
			store(ts)
		case "erase":
			erase(ts)
		default:
			usage()
			os.Exit(1)
		}
		sync(ts)
	} else {
		usage()
	}
}
