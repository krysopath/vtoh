# usage

A simple vault token helper written in golang. Do not use, just put in $PATH: Vault will use it. Token helpers are meant to help with storage and retrieval of vault tokens. 

> This one helper stores valid token in a per VAULT_ADDR keyed map, making working with many VAULT_ADDR a comforting reality..

# installation

```
$ make
go get
go mod graph > deps.txt
go build -ldflags='-s -w -X main.gitTag=0.0.0 -X main.gitRef=3e79321' -o build/bin/token-helper
sudo cp build/bin/token-helper /usr/bin/token-helper
```

# configuration

Create file at `~/.vault` with this content:
```
token_helper = "/usr/bin/token-helper"
```

# outlook

- crypto layer with gpg for `~/.vault-tokens` file

