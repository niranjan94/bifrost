### Bifrost

> WIP

##### Prerequisites
- Configured AWS CLI environment with access credentials. ([How to](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html))
- Installed docker environment. ([How to](https://docs.docker.com/install/))

##### Installation

```bash
curl https://latest-release.now.sh/niranjan94/bifrost/$(uname -s)_$(uname -m) -L -o bifrost.tar.gz
sudo tar --directory /usr/bin/ -xzf bifrost.tar.gz bifrost
```

##### Usage

```
Usage:
  bifrost [command]

Available Commands:
  deploy      Deploy your stack to the cloud
  help        Help about any command

Flags:
  -c, --config string    config file (default is ./bifrost.yaml)
  -d, --dry-run          dry run mode (default is false)
      --functions-only   Deploy only functions
  -h, --help             help for bifrost
      --only string      Deploy only specific resources
  -r, --region string    region (default is ap-southeast-1) (default "ap-southeast-1")
  -s, --stage string     Stage to use (default is dev) (default "dev")
  -v, --verbose          Verbose mode

Use "bifrost [command] --help" for more information about a command.
```