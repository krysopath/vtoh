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
	VAULT_ADDR string = os.Getenv("VAULT_ADDR")
	hashedAddr string = hash(VAULT_ADDR)
	gitRef     string = ""
	gitTag     string = ""
)

type Backend interface {
	Load() ([]byte, error)
	Save([]byte) (bool, error)
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
			Region: "eu-central-1",
			Path:   source.Path}
	default:
	}
	ts.Backend = backend

	content, _ := ts.Backend.Load()
	yaml.Unmarshal(content, &ts.Data)
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`%s %s-%s
	
	important: this token helper is not meant to be executed directly
	supported commands: get, store, erase
`, os.Args[0], gitTag, gitRef,
	)
}

func main() {
	if len(VAULT_ADDR) == 0 {
		fmt.Fprintln(os.Stderr, "err: VAULT_ADDR is unset")
		os.Exit(100)
	}
	if len(os.Args) >= 2 {
		ts := &TokenStore{
			DataSourceURI: fmt.Sprintf("s3://%s", s3Object),
		}
		ts.Init()
		switch os.Args[1] {
		case "get":
			token, _ := ts.Data[hashedAddr].(string)
			fmt.Fprintf(os.Stdout, "%s", token)
		case "store":
			reader := bufio.NewReader(os.Stdin)
			token, _ := reader.ReadString('\n')
			ts.Data[hashedAddr] = token
		case "erase":
			delete(ts.Data, hashedAddr)
		default:
			usage()
			os.Exit(1)
		}
		storeBytes, err := yaml.Marshal(ts.Data)
		if err != nil {
			panic(err)
		}
		ts.Backend.Save(storeBytes)
	} else {
		usage()
	}
}
