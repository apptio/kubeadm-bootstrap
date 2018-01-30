package cmd

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "kubeadm.libsonnet",
		FileModTime: time.Unix(1517329782, 0),
		Content:     string("{\n    // Required arguments for this template\n    k8sVersion:: \"v1.8.4\",\n    clusterName:: std.extVar(\"clustername\"),\n    addressList:: std.split(std.extVar(\"addresslist\"), \",\"),\n\n    local k8sVersion = $.k8sVersion,\n\n    local clusterName = $.clusterName,\n\n    local datacenterName = std.extVar(\"datacenter\"),\n\n    local domainName = std.extVar(\"domainname\"),\n\n    local bootstrapMasterNodeName = std.extVar(\"nodename\"),\n\n    local cloudProvider = std.extVar(\"cloudprovider\"),\n\n    local ipAddress = std.extVar(\"ipaddress\"),\n\n    local token = std.extVar(\"token\"),\n\n    local apiServerExtraArgs = {\n        \"etcd-prefix\": datacenterName + \"-\" + clusterName,\n        profiling: \"false\",\n        //\"admission-control\": \"Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota,AlwaysPullImages,DenyEscalatingExec,SecurityContextDeny,PodSecurityPolicy\",\n        \"audit-log-path\": \"-\",\n        \"audit-log-maxage\": \"30\",\n        \"audit-log-maxbackup\": \"10\",\n        \"audit-log-maxsize\": \"100\",\n        \"service-account-lookup\": \"true\",\n        \"repair-malformed-updates\": \"false\",\n        \"cloud-provider\": cloudProvider,\n        \"advertise-address\": ipAddress,\n    },\n\n    local controllerManagerExtraArgs = {\n        profiling: \"false\",\n        \"cloud-provider\": cloudProvider,\n    },\n\n    local schedulerExtraArgs = {\n        profiling: \"false\",\n    },\n\n    local etcdCount = 3,\n\n    local etcdEndpoints = std.makeArray(3, function(count) \"https://\" + datacenterName + \"-\" + clusterName + \"etcd\" + \"-\" + std.toString(count + 1) + \".\" + domainName + \":2379\"),\n\n    local apiServerIPs = $.addressList,\n\n    local apiServerNames = std.makeArray(3, function(count) datacenterName + \"-\" + clusterName + \"master\" + \"-\" + std.toString(count + 1) + \".\" + domainName),\n\n    local apiServerDiscoveryNames = [\n        datacenterName + \"-\" + clusterName + \"master\" + \".\" + domainName,\n        clusterName + \".service.discover\",\n        datacenterName + \"-\" + clusterName + \".service.discover\",\n        datacenterName + \"-\" + clusterName + \".\" + datacenterName + \".service.discover\",\n    ],\n\n    //local apiServerCNAME = $.datacenterName + \"-\" + clusterName + \"master\" + \".\" + $.domainName,\n\n    local apiServerCertSANs = [apiServerNames, apiServerIPs, apiServerDiscoveryNames],\n    local etcd = true,\n\n    apiVersion: \"kubeadm.k8s.io/v1alpha1\",\n    kind: \"MasterConfiguration\",\n    kubernetesVersion: k8sVersion,\n    nodeName: bootstrapMasterNodeName,\n    tokenTTL: \"0\",\n    token: token,\n    api: {\n        advertiseAddress: \"0.0.0.0\",\n    },\n    apiServerExtraArgs: apiServerExtraArgs,\n    controllerManagerExtraArgs: controllerManagerExtraArgs,\n    schedulerExtraArgs: schedulerExtraArgs,\n    apiServerCertSANs: std.flattenArrays(apiServerCertSANs),\n    cloudProvider: cloudProvider,\n    etcd: {\n        [if etcd then \"endpoints\"]: etcdEndpoints,\n        [if etcd then \"caFile\"]: \"/etc/kubernetes/puppet/ca.pem\",\n        [if etcd then \"certFile\"]: \"/etc/kubernetes/puppet/cert.pem\",\n        [if etcd then \"keyFile\"]: \"/etc/kubernetes/puppet/key.pem\",\n    },\n\n\n}\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1517329782, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "kubeadm.libsonnet"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`../lib`, &embedded.EmbeddedBox{
		Name: `../lib`,
		Time: time.Unix(1517329782, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"kubeadm.libsonnet": file2,
		},
	})
}
