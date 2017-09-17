# awshostname - Retrieve Public DNS name of EC2 by tags

## Introduction
When you launch EC2 instances, you will attach some tags on it. You can "describe" them and filter by tags to obtain public hostname to do ssh.
This tool helps you to set up ~/.ssh/config.

## Usage

```
Usage: awshostname [options] <hostname>
  -p string
        Profile to use (default "default")
  -r string
        AWS region to search in
  -t string
        Tags to filter the instances. TAG=VALUE,TAG=VALUE...
```

If your host has "Name=testinstance" tag. Then you may put the following into your ~/.ssh/config:
```
Host testinstance
  User your_user
  IdentityFile ~/.ssh/your_key
  ProxyCommand nc `awshostname -t Name=testinstance %h` %p
```

Then you can ssh into the instance by:
```
ssh testinstance
```

## Advanced
You can give two or more tags on the commandline separated by commas.
```
awshostname -t tag1=val1,tag2=val2 hostname
```

If you have two or more hosts with the same tag set, you can specify it by adding "#" + index after the first component. The index starts with 0.
```
awshostname -t tag1=val1 hostname#2.domain  # The third host of val1
```

-t and -r parameters can take index of the hostname components with "#":
```
# tag1 has the value "hoge"
awshostname -t tag1=#2 foo.hoga.hoge
```

## License
See [LICENSE].
