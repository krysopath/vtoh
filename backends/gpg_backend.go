//// +build backends/gpg

package backends

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"golang.org/x/crypto/openpgp"
	"gopkg.in/yaml.v2"
)

var (
	rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func getKeyByEmail(keyring openpgp.EntityList, keyInfo string) *openpgp.Entity {
	if len(keyInfo) > 254 || !rxEmail.MatchString(keyInfo) {
		fmt.Printf("err: %s is not a valid email address\n", keyInfo)
		return nil
	}
	for _, entity := range keyring {
		for _, ident := range entity.Identities {
			if ident.UserId.Email == keyInfo {
				return entity
			}
		}
	}
	return nil
}

type GpgBackend struct {
	FilePath    string
	Recipients  []string
	KeyRingHome string
	Signer      string
	Password    string
}

func (backend GpgBackend) PubKeyRing() string {
	return fmt.Sprintf("%s/pubring.gpg", backend.KeyRingHome)
}
func (backend GpgBackend) PrivKeyRing() string {
	return fmt.Sprintf("%s/secring.gpg", backend.KeyRingHome)
}

func (backend GpgBackend) Encrypt(data []byte) string {
	keyringFileBuffer, _ := os.Open(backend.PubKeyRing())
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		panic(err)
	}
	recipients := []*openpgp.Entity{}
	for _, recipient := range backend.Recipients {
		recipients = append(
			recipients,
			getKeyByEmail(entityList, recipient),
		)
	}

	buf := new(bytes.Buffer)
	w, err := openpgp.Encrypt(buf, recipients, nil, nil, nil)
	if err != nil {
		panic(err)
	}
	_, err = w.Write([]byte(data))
	if err != nil {
		panic(err)
	}
	err = w.Close()
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(buf)
	if err != nil {
		panic(err)
	}
	encStore := base64.StdEncoding.EncodeToString(bytes)

	return encStore
}

func (backend GpgBackend) Decrypt(data []byte) []byte {
	keyringFileBuffer, _ := os.Open(backend.PrivKeyRing())
	defer keyringFileBuffer.Close()
	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		panic(err)
	}
	recipients := []*openpgp.Entity{}
	for _, recipient := range backend.Recipients {
		recipients = append(
			recipients,
			getKeyByEmail(entityList, recipient),
		)
	}

	entity := recipients[0]
	entity.PrivateKey.Decrypt([]byte(``))
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt([]byte(``))
	}
	dec, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		panic(err)
	}

	// Decrypt it with the contents of the private key
	md, err := openpgp.ReadMessage(bytes.NewBuffer(dec), entityList, nil, nil)
	if err != nil {
		panic(err)
	}
	dataBytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		panic(err)
	}
	return dataBytes

}

func (backend GpgBackend) Load() ([]byte, error) {
	content := make([]byte, 0)
	if _, err := os.Stat(backend.FilePath); err == nil {
		encContent, err := ioutil.ReadFile(backend.FilePath)
		if err != nil {
			panic(err)
		}
		content = backend.Decrypt(encContent)
	}
	return content, nil
}

func (backend GpgBackend) Save(data interface{}) (bool, error) {
	dataBytes, yamlErr := yaml.Marshal(data)
	if yamlErr != nil {
		panic(yamlErr)
	}
	encBytes := backend.Encrypt(dataBytes)

	writeErr := ioutil.WriteFile(
		backend.FilePath,
		[]byte(encBytes),
		0600)

	if writeErr != nil {
		panic(writeErr)
	}
	return true, nil
}
