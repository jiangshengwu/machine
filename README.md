# Docker Machine

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

It works a bit like this:

```console
$ docker-machine create --driver aliyun\
                        --aliyun-access-key-id id\
                        --aliyun-access-key-secret secret\
                        --aliyun-region-id cn-beijing\
                        --aliyun-image-id centos6u5_64_20G_aliaegis_20150130.vhd\
                        --aliyun-security-group-id sg-25mnyu5qm\
                        --aliyun-docker-args="--ip-forward=false --bridge=none --iptables=false -s=devicemapper"\
                        --aliyun-docker-registry registry.mirrors.aliyuncs.com\
                        centos
Creating SSH Key Pair...
Creating ECS instance...
Allocating public IP address...
Starting instance, this may take several minutes...
Upload SSH key to machine...
Created Instance ID i-2576oxvs6, Public IP address 182.92.148.199, Private IP address 10.171.133.223
Updating Metadata of yum...
Run yum install for curl...
Installing Docker from yum...
Updating Metadata of yum...
Run yum install for docker-io...
Try to start Docker daemon...
To see how to connect Docker to this machine, run: machine env centos

$ eval $(docker-machine env centos)

$ docker-machine ls
NAME     ACTIVE   DRIVER   STATE     URL                         SWARM
aliyun            aliyun   Running   tcp://182.92.243.188:2376
centos   *        aliyun   Running   tcp://182.92.148.199:2376
ubuntu            aliyun   Running   tcp://123.56.93.119:2376

$ docker run busybox echo hello world
Unable to find image 'busybox:latest' locally
Pulling repository busybox
8c2e06607696: Download complete
cf2616975b4a: Download complete
6ce2e90b0bc7: Download complete
Status: Downloaded newer image for busybox:latest
hello world
```

## Aliyun Options

* `--driver` Set to *aliyun*
* `--aliyun-access-key-id` (Required) ECS Access Key Id
* `--aliyun-access-key-secret` (Required) ECS Access Key Secret
* `--aliyun-region-id` (Required) ECS Region: *cn-beijing*, etc.
* `--aliyun-image-id` (Optional) ECS Image Id, set to *ubuntu1404_64_20G_aliaegis_20150325.vhd* as default if not given
* `--aliyun-instance-type-id` (Optional) ECS Instance Type Id, set to *ecs.t1.small* as default if not given
* `--aliyun-security-group-id` (Optional) ECS Security Group Id, a new group will be created if not given
* `--aliyun-ssh-pass` (Optional) ECS SSH Password, set to *ASDqwe123* as default if not given
* `--aliyun-bandwidth-out` (Optional) ECS Internet Out Bandwidth(MB), set to *1* as default if not given
* `--aliyun-docker-args` (Optional, only supports CentOS6.x Image) Docker Daemon Args: *--ip-forward=false --bridge=none --iptables=false -s=devicemapper*, etc
* `--aliyun-docker-registry` (Optional, only supports CentOS6.x Image) Docker Insecure Registry: *registry.mirrors.aliyuncs.com*, etc

## Installation and documentation

Full documentation [is available here](https://docs.docker.com/machine/).

## Contributing

Want to hack on Machine? Please start with the [Contributing Guide](https://github.com/docker/machine/blob/master/CONTRIBUTING.md).