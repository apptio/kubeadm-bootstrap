# kubeadm-bootstrap

[![Build Status](https://travis-ci.org/apptio/kubeadm-bootstrap.svg?branch=master)](https://travis-ci.org/apptio/kubeadm-bootstrap)

kubeadm-bootstrap is a simple tool to generate [kubeadm](https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/) configuration files.

It uses [jsonnet](http://jsonnet.org/) to template the resulting JSON formatted kubeadm configuration file, meaning you can easily generate configurations for kubeadm, which can then be passed to a configuration management tool like [Puppet](https://puppet.com/) or [Chef](https://www.chef.io/chef/)

## Rationale

If you're managing lots of Kubernetes clusters, you want to make sure they are uniform, and adhere to a set of known good standards. Kubeadm is good for actually bootstrapping the Kubernetes clusters, but editing and creating the configuration files involves some degree of templating and management.

kubeadm-bootstrap is designed to take the pain out of this process. It's a very opinionated tool, and will generate a configuration with quite a few assumptions. Please see the [assumptions](#assumptions) section for more information.

## Usage

kubeadm-bootstrap will attempt to detect as many defaults as it can. It will try and automatically detect the domain name, hostname, datacenter (using Puppet facter facts) as well as everything else. You can override all of this using paramaters:

```base
Generate a kubeadm config for a kubernetes cluster using CIS compatible configuration
using jsonnet templates for the config file

Usage:
  kubeadm-bootstrap [flags]
  kubeadm-bootstrap [command]

Available Commands:
  help        Help about any command
  version     return the current version of kubeadm-bootstrap

Flags:
  -a, --addresslist string   comma separated list of IP's for the cluster
  -c, --clustername string   cluster name for cluster bootstrap (default "k1")
      --config string        config file (default is $HOME/.kubeadm-bootstrap.yaml)
  -d, --datacenter string    datacenter name for cluster boostrap
  -D, --domainname string    domain name for nodes in cluster
      --dry-run              output the kubeadm config to stdout instead of a file
  -h, --help                 help for kubeadm-bootstrap
  -f, --kubeadmfile string   path to kubeadm file to write (default "/etc/kubernetes/kubeadm.json")
  -n, --nodename string      nodename for bootstrap master
  -m, --number int           number of masters in the cluster (default 3)
  -s, --svcip string         kubernetes service IP (default "10.96.0.1")
  -t, --token string         kubernetes bootstrap token

Use "kubeadm-bootstrap [command] --help" for more information about a command.
```

## Installation

Coming Soon

## Assumptions

There are quite a lot of assumptions when using kubeadm-bootstrap, so please use it with caution. Many of these assumptions will be fixed as the tool is developed - pull requests are welcome!

These assumptions includes:

- Etcd is external from your cluster
- Etcd uses TLS
- The TLS certifcates for your cluster like in `/etc/kubernetes/puppet`
- You have a service discovery domain of `service.discover` (We use [consul](https://consul.io))
- The kubernetes clusters are numbered/named using the convention `k{1,2,3}` per datacenter. By default your cluster will be named `k1`
- The naming convention for your masters is something like `${datacenter}-${clustername}master-{master_number}.${domain}`

## Building

See the [docs](docs/BUILDING.md)

## Contributing

Fork the repo and send a merge request! See the issues for more information.

# Caveats

There are currently no tests, and the code is not very [DRY](https://en.wikipedia.org/wiki/Don%27t_repeat_yourself).

This was one of Apptio's first exercises in Go, and pull requests are very welcome.
