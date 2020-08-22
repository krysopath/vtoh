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

#### Creating a Key
```
$ gpg --full-generate-key
gpg (GnuPG) 2.2.19; Copyright (C) 2019 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
  (14) Existing key from card
Your selection? 1
RSA keys may be between 1024 and 4096 bits long.
What keysize do you want? (3072) 4096
Requested keysize is 4096 bits
Please specify how long the key should be valid.
         0 = key does not expire
      <n>  = key expires in n days
      <n>w = key expires in n weeks
      <n>m = key expires in n months
      <n>y = key expires in n years
Key is valid for? (0) 0
Key does not expire at all
Is this correct? (y/N) y

GnuPG needs to construct a user ID to identify your key.

Real name: special dude
Email address: sepcial@dude.tld
Comment: very special indeed
You selected this USER-ID:
    "special dude (very special indeed) <sepcial@dude.tld>"

Change (N)ame, (C)omment, (E)mail or (O)kay/(Q)uit? O
We need to generate a lot of random bytes. It is a good idea to perform
some other action (type on the keyboard, move the mouse, utilize the
disks) during the prime generation; this gives the random number
generator a better chance to gain enough entropy.
We need to generate a lot of random bytes. It is a good idea to perform
some other action (type on the keyboard, move the mouse, utilize the
disks) during the prime generation; this gives the random number
generator a better chance to gain enough entropy.
gpg: key E19A643174D10D65 marked as ultimately trusted
gpg: revocation certificate stored as '/home/gve/.gnupg/openpgp-revocs.d/EA0D97C8C74A7815710E84E2E19A643174D10D65.rev'
public and secret key created and signed.

pub   rsa4096 2020-08-22 [SC]
      EA0D97C8C74A7815710E84E2E19A643174D10D65
uid                      special dude (very special indeed) <special@dude.tld>
sub   rsa4096 2020-08-22 [E]
```
> A current limitation is impossibility to send gpg passphrase to token-helper. Create a key without a passphrase for now.

export the key:

```
gpg --export-secret-keys --out gnupg/secring.gpg EA0D97C8C74A7815710E84E2E19A643174D10D65
gpg --export --out gnupg/pubring.gpg  EA0D97C8C74A7815710E84E2E19A643174D10D65
gpg --export-secret-keys --out gnupg/secring.gpg special

export VAULT_TOKEN_SRC="gpg:///$HOME/.vault-tokens.asc?&recipients=special@dude.tld"
```

> Note: the recipient url query can be specified multiple times. Each recipient
>must exist in the keyring of token-helper.

```
$ token-helper config
backend:
  filepath: /home/gve/.vault-tokens.asc
  recipients:
  - special@dude.tld
  keyringhome: /home/gve/.token-helper
source_uri: gpg:///home/gve/.vault-tokens.asc?&recipients=special@dude.tld
data: {}
```

> Remember that we do not support new kbx formatted key rings yet and that no gpg-agemnt integration exists.

From here on buisness as usual:
```
$ vault login $(cat ~/.vault-token)
Success! You are now authenticated. The token information displayed below
is already stored in the token helper. You do NOT need to run "vault login"
again. Future Vault requests will automatically use this token.

Key                  Value
---                  -----
token                s.*********************
token_accessor       q**********************o
token_duration       ∞
token_renewable      false
token_policies       ["admin"]
identity_policies    []
policies             ["admin"]
gve@stateless:~/src/token-helper$ token-helper config
backend:
  filepath: /home/gve/.vault-tokens.asc
  recipients:
  - special@dude.tld
  keyringhome: /home/gve/.token-helper
source_uri: gpg:///home/gve/.vault-tokens.asc?&recipients=special@dude.tld
data:
  00fc7f3ebd385708908f63a377127b8e1a47980b8be30b7e79d3fd71f34aa5dd: s.*************************

```

### S3Backend

- use aws/s3 buckets
- allows for ACL via IAM
- serverside KMS
- might allow user defined KMS keys in the future


# outlook

- dbus backend
- gpg-agent integration

