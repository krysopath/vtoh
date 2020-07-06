# usage

A simple vault token helper written in golang. Do not use, just put in $PATH:
Vault will use it. Token helpers are meant to help with storage and retrieval
of vault tokens. 

> This one helper stores valid token in a per VAULT_ADDR keyed map, making
> working with many VAULT_ADDR a comforting reality..

# installation

```
$ make
go get
go mod graph > deps.txt
go build -ldflags='-s -w -X main.gitTag=0.0.0 -X main.gitRef=3e79321' -o build/bin/token-helper
sudo cp build/bin/token-helper /usr/bin/token-helper
```
## configuration

Create file at `~/.vault` with this content:
```
token_helper = "/usr/bin/token-helper"
```

> `vault` will use the helper binary from now on

Currently supported:
- `FileBackend`,  `VAULT_TOKEN_SRC=file:///path/to/file`
- `S3Backend`, `VAULT_TOKEN_SRC=s3://path/to/object`
- `GpgBackend`, `VAULT_TOKEN_SRC=gpg://path/to/file`

To select a backend, set `VAULT_TOKEN_SRC`:
```
export VAULT_TOKEN_SRC=s3://personal-secret-bucket/tokens
```
> e.g. would use the bucket `personal-secret-bucket` and `tokens` as storage object

Use `file://` to load the store from a local file:
```
export VAULT_TOKEN_SRC=file://$HOME/.vault-tokens
```

> You will need to issue a `vault login` now

### FileBackend

- stores store in a definable file
- plaintext
- no dependencies

> This file-based backend works very simple. 

### GpgBackend

- uses gpg crypto to secure the token storage file
- saves store as friendly base64 string
- does not support interactive password prompts (vault itself does not allow working with stdin/out)
- might integrate with gpg-agent in the future

> The gold standard of crypto. We can do at least this.

### S3Backend

- use aws/s3 buckets
- allows for ACL via IAM
- serverside KMS
- might allow user defined KMS keys in the future


# outlook

- dbus backend
- gpg-agent integration

