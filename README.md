# awshostname - Retrieve Public DNS name of EC2 by tags

## Introduction
When you launch EC2 instances, you will attach some tags on it. You can "describe" them and filter by tags to obtain public hostname to do ssh.
However it is often cumbersome to find specific host with certain tags.
This tool is simply designed for finding EC2 host with specific tags. You can use it in ~/.ssh/config to make variety of hosts accessible with small set of Host/ProxyCommand rules.

## Usage

```
awshostname -h
Usage: ./awshostname [options] [HostSpec]
  -d    Show debug information
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
  ProxyCommand nc `awshostname -t Name=testinstance` %p
```

Then you can ssh into the instance by:
```
ssh testinstance
```

You can give two or more tags on the commandline separated by commas.
```
awshostname -t tag1=val1,tag2=val2
```

## Advanced
-t and -r parameters can take index of the HostSpec components with "#":
```
# Name has the value "foo"
awshostname -t Name=#0 foo.hoga.hoge
```

For example you can put the following in your ~/.ssh/config:
```
Host *.mydomain
  User your_user
  IdentityFile ~/.ssh/your_key
  ProxyCommand nc `awshostname -t Name=#0 %h` %p
```
to ssh into any host with Name tag by adding ".mydomain" suffix.

If you have two or more hosts with the same tag set, you can specify it by adding "#" + index after the first component. The index starts with 0. The matching instances are sorted by LaunchTime.
```
# The third host which has Name=foo
awshostname -t Name=#0 foo#2.fuga.hoge
```

## License
This software is licensed under [MIT license](LICENSE).
