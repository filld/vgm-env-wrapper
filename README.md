# vgm-env-wrapper

Wrapper for [VGM](https://github.com/ChannelMeter/vault-gatekeeper-mesos) secrets in Mararthon. It replaces environment variables values with the data fetched from vault.

## Usage

Set environment variables:
```
VAULT_ADDR=http://localhost:8200
GATEKEEPER_ADDR=http://localhost:9201
VGM_ENV_ENABLED=1
```

if `VGM_ENV_ENABLED` is not set, `vgm-env-wrapper` will not do anything with environment variables.

When you run it under Mesos/Marathon `MESOS_TOKEN` variable will contain Mesos task ID. It will be automatically used to fetch
token from VGM.

Define required variables in one of two formats:
* `SECRET_TOKEN=vgm:secret/test:key`. Secret with path 'secret/test' will be fetched from vault, and environment variable value will be replaced with the secret data with key `key`.
* `SECRET_FILE_PATH=vgm_file:secret/test:key`. Secret with path 'secret/test' will be fetched from vault, secret data with key `key` will be saved to a temporary file, those path will replace the environment variable value.

Call the script like this:
```
vgm-env-wrapper [command with args]
```

### Docker
For usage inside docker, just put this script as an ENTRYPOINT

## Author
[Gleb Pomykalov](http://github.com/glebpom)
