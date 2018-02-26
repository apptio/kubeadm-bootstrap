{

    string_to_int(s)::
        local char_to_int(c) = std.codepoint(c) - std.codepoint("0");
        local digits = std.map(char_to_int, std.stringChars(s));
        std.foldr(function(x, y) x + y,
                  std.makeArray(std.length(digits),
                                function(x) digits[std.length(digits) - x - 1] * std.pow(10, x)),
                  0),


    // Required arguments for this template
    k8sVersion:: "v1.8.4",
    clusterName:: std.extVar("clustername"),
    addressList:: std.split(std.extVar("addresslist"), ","),

    local k8sVersion = $.k8sVersion,

    local clusterName = $.clusterName,

    local datacenterName = std.extVar("datacenter"),

    local domainName = std.extVar("domainname"),

    local bootstrapMasterNodeName = std.extVar("nodename"),

    local cloudProvider = std.extVar("cloudprovider"),

    local ipAddress = std.extVar("ipaddress"),

    local token = std.extVar("token"),

    local numberMasters = std.extVar("number_masters"),

    local apiServerExtraArgs = {
        "etcd-prefix": datacenterName + "-" + clusterName,
        profiling: "false",
        //"admission-control": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota,AlwaysPullImages,DenyEscalatingExec,SecurityContextDeny,PodSecurityPolicy",
        "audit-log-path": "-",
        "audit-log-maxage": "30",
        "audit-log-maxbackup": "10",
        "audit-log-maxsize": "100",
        "service-account-lookup": "true",
        "repair-malformed-updates": "false",
        "apiserver-count": numberMasters,
        "cloud-provider": cloudProvider,
        "advertise-address": ipAddress,
        "request-timeout": "300s",
        "admission-control": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,PersistentVolumeLabel,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,ResourceQuota,AlwaysPullImages,DenyEscalatingExec,SecurityContextDeny",
    },

    local controllerManagerExtraArgs = {
        profiling: "false",
        "terminated-pod-gc-threshold": "10",
        "cloud-provider": cloudProvider,
        "address": "0.0.0.0",
    },

    local schedulerExtraArgs = {
        profiling: "false",
        "address": "0.0.0.0",
    },

    local etcdCount = 3,

    local etcdEndpoints = std.makeArray(3, function(count) "https://" + datacenterName + "-" + clusterName + "etcd" + "-" + std.toString(count + 1) + "." + domainName + ":2379"),

    local apiServerIPs = $.addressList,

    local apiServerNames = std.makeArray($.string_to_int(numberMasters), function(count) datacenterName + "-" + clusterName + "master" + "-" + std.toString(count + 1) + "." + domainName),

    local apiServerDiscoveryNames = [
        datacenterName + "-" + clusterName + "master" + "." + domainName,
        clusterName + ".service.discover",
        datacenterName + "-" + clusterName + ".service.discover",
        datacenterName + "-" + clusterName + "." + datacenterName + ".service.discover",
    ],

    //local apiServerCNAME = $.datacenterName + "-" + clusterName + "master" + "." + $.domainName,

    local apiServerCertSANs = [apiServerNames, apiServerIPs, apiServerDiscoveryNames],
    local etcd = true,

    apiVersion: "kubeadm.k8s.io/v1alpha1",
    kind: "MasterConfiguration",
    kubernetesVersion: k8sVersion,
    nodeName: bootstrapMasterNodeName,
    tokenTTL: "0",
    token: token,
    api: {
        advertiseAddress: "0.0.0.0",
    },
    apiServerExtraArgs: apiServerExtraArgs,
    controllerManagerExtraArgs: controllerManagerExtraArgs,
    schedulerExtraArgs: schedulerExtraArgs,
    apiServerCertSANs: std.flattenArrays(apiServerCertSANs),
    cloudProvider: cloudProvider,
    etcd: {
        [if etcd then "endpoints"]: etcdEndpoints,
        [if etcd then "caFile"]: "/etc/kubernetes/puppet/ca.pem",
        [if etcd then "certFile"]: "/etc/kubernetes/puppet/cert.pem",
        [if etcd then "keyFile"]: "/etc/kubernetes/puppet/key.pem",
    },


}
