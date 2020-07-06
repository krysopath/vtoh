package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"net/url"
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
	s3Object string = fmt.Sprintf(
		"%s/%s",
		"tf-stage-internal-vault-storage",
		".vault-tokens",
	)
	VAULT_ADDR      string = os.Getenv("VAULT_ADDR")
	VAULT_TOKEN_SRC string = os.Getenv("VAULT_TOKEN_SRC")
	hashedAddr      string = hash(VAULT_ADDR)
	gitRef          string = ""
	gitTag          string = ""
)

type Backend interface {
	Load() ([]byte, error)
	Save(interface{}) (bool, error)
}

type TokenStore struct {
	Backend       Backend                `yaml:"backend"`
	DataSourceURI string                 `yaml:"source_uri"`
	Data          map[string]interface{} `yaml:"data"`
}

func (ts *TokenStore) Init() {
	source, err := url.Parse(ts.DataSourceURI)
	if err != nil {
		panic(err)
	}

	var backend Backend

	switch source.Scheme {
	case "file":
		backend = FileBackend{FilePath: source.Path}
	case "s3":
		backend = S3Backend{
			Bucket: source.Host,
			//Region: "eu-central-1",
			Path: source.Path}
	default:
	}
	ts.Backend = backend

	content, _ := ts.Backend.Load()
	yaml.Unmarshal(content, &ts.Data)
}

func config(data interface{}) {
	configBytes, _ := yaml.Marshal(data)
	fmt.Fprintf(os.Stderr, `%s`, configBytes)
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`%s %s-%s
	important: this token helper is not meant to be executed directly
	supported commands: get, store, erase & config

	where only config is intended to invoked interactively
`, os.Args[0], gitTag, gitRef,
	)
}

func main() {
	if len(VAULT_ADDR) == 0 {
		fmt.Fprintln(os.Stderr, "err: VAULT_ADDR is unset")
		os.Exit(100)
	}
	if len(VAULT_TOKEN_SRC) == 0 {
		fmt.Fprintln(os.Stderr, "err: VAULT_TOKEN_SRC is unset")
		os.Exit(101)
	}
	if len(os.Args) >= 2 {
		store := &TokenStore{
			DataSourceURI: VAULT_TOKEN_SRC,
		}
		store.Init()
		switch os.Args[1] {
		case "get":
			token, _ := store.Data[hashedAddr].(string)
			fmt.Fprintf(os.Stdout, "%s", token)
		case "store":
			reader := bufio.NewReader(os.Stdin)
			token, _ := reader.ReadString('\n')
			store.Data[hashedAddr] = token
		case "erase":
			delete(store.Data, hashedAddr)
		case "config":
			config(store)
		default:
			usage()
			os.Exit(1)
		}
		store.Backend.Save(store.Data)
	} else {
		usage()
	}
}
