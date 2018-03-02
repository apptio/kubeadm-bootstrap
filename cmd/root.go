// Copyright © 2018 Lee Briggs <lee@leebriggs.co.uk>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/GeertJohan/go.rice"
	jsonnet "github.com/google/go-jsonnet"

	log "github.com/Sirupsen/logrus"

	"io/ioutil"

	//"net"
	"strings"

	n "github.com/apptio/kubeadm-bootstrap/pkg/net"
	t "github.com/apptio/kubeadm-bootstrap/pkg/token"
)

var cfgFile string
var nodeName string
var clusterName string
var datacenter string
var domainName string
var kubeadmFile string
var cloudProvider string
var hostname string
var ipAddress string
var detectedDomainName string
var addressList string
var addresses string
var numberMasters int
var svcIP string
var token string
var generatedToken string
var dryrun bool

// Version string
var Version string

const libLocation = "../lib"
const kubeSonnet = "kubeadm.libsonnet"
const awsCloud = "aws"
const awslocalhostname = "local-hostname"
const awsDN = ".compute.internal"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubeadm-bootstrap",
	Short: "Bootstrap a kubernetes cluster using known good config",
	Long: `Generate a kubeadm config for a kubernetes cluster using CIS compatible configuration
using jsonnet templates for the config file`,
	Run: func(cmd *cobra.Command, args []string) {
		kb := NewKubeBootstrap(datacenter, clusterName)
		kubeBootstrap(kb)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	Version = version
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func kubeBootstrap(kb *KubeBootstrap) {

	template, err := kb.getRiceTemplate(libLocation, kubeSonnet)

	if kb.DCName == "" {
		kb.DCName = kb.getDatacenterFromFacter()
		if kb.DCName == "" {
			// we still cant find the datacenter and none was passed in, BAIL
			log.Fatalln("No datacenter information was provided, and we cannot detect the datacenter. Exiting.")
			return
		}
	}

	if kb.Cluster == "" {
		log.Fatalln("No cluster information provided. Exiting.")
		return
	}

	// TODO: first pass at the refactor
	isAWS := kb.getAWSInformation("default")

	// TODO: this is kind of ugly but this is a first pass,
	// TODO: need to modularize before adding in the DI scaffolding
	if !isAWS {
		log.Info("Not running in AWS")
		cloudProvider = ""
		hostname, err = os.Hostname()
		if err != nil {
			log.Fatal("Cannot detect hostname", err)
		}
		splitHostname := strings.Split(hostname, ".")
		if len(splitHostname) < 3 {
			log.Warn("Cannot auto detect domainname")
		} else {
			detectedDomainName = splitHostname[1] + "." + splitHostname[2]
		}
		kb.DomainName = detectedDomainName
	}

	if nodeName == "" {
		if hostname == "" {
			log.Fatalln("Unable to detect hostname and no hostname provided")
			return
		}

		// TODO: try autodetect hostname
		log.Info("No hostname provided - auto detecting hostname")
		nodeName = hostname
	}

	if addressList == "" {
		addresses = n.GetMasterAddresses(kb.DCName, kb.Cluster, kb.DomainName, numberMasters, svcIP)
	} else {
		addresses = addressList
	}

	ipAddress = n.GetOutboundIP()

	if token == "" {

		generatedToken, err = t.GenerateToken()

		if err != nil {
			log.Fatal("Error generating bootstrap token", err)
		}
		token = generatedToken
	}

	// create a jsonnet vm
	vm := jsonnet.MakeVM()
	// populate jsonnet extvars
	vm.ExtVar("datacenter", kb.DCName)
	vm.ExtVar("clustername", kb.Cluster)
	vm.ExtVar("domainname", kb.DomainName)
	vm.ExtVar("nodename", nodeName)
	vm.ExtVar("cloudprovider", cloudProvider)
	vm.ExtVar("ipaddress", ipAddress)
	vm.ExtVar("addresslist", addresses)
	vm.ExtVar("token", token)
	vm.ExtVar("number_masters", strconv.Itoa(numberMasters))

	// evaluate jsonnet snippet
	out, err := vm.EvaluateSnippet("file", template)

	if err != nil {
		log.Fatal(err)
	}

	if !dryrun {
		// write the kubeadm file to disk
		outFile := []byte(out)
		err = ioutil.WriteFile(kubeadmFile, outFile, 0644)

		if err != nil {
			log.Fatal("Error writing kubeadm file", err)
		}

		log.Info("Wrote kubeadm file: ", kubeadmFile)
	} else {
		log.Info("Dry run specified, printing to stdout: ")
		fmt.Println(out)
	}
}

func (kb *KubeBootstrap) getRiceTemplate(fLocation string, fName string) (string, error) {
	// read static assets
	var template string
	templateBox, err := rice.FindBox(fLocation)
	if err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"location": fLocation,
		}).Errorln("Error finding ricebox location")
		return template, err
	}

	template, err = templateBox.String(fName)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"file":  fName,
		}).Errorln("Error finding rice box location")
		return template, err
	}

	log.WithFields(log.Fields{
		"template": template,
	}).Debugln("Found rice box template")

	return template, nil
}

func (kb *KubeBootstrap) getDatacenterFromFacter() string {
	log.Infoln("Trying to autodetect datacenter name")
	var dcName string

	out, err := exec.Command("facter", "-p", "datacenter").Output()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorln("Error finding datacenter from facter")
		return dcName
	}

	dcName = strings.TrimSpace(string(out))

	if dcName == "" {
		log.Errorln("Could not detect datacenter from facter")
	} else {
		log.WithFields(log.Fields{
			"datacenter": dcName,
		}).Infoln("Found datacenter from facter")
	}

	return dcName
}

func (kb *KubeBootstrap) getAWSInformation(profile string) bool {
	// determine if we're in AWS:

	sess, err := kb.CreateAWSSession(profile)

	if err != nil {
		log.Fatal("Error creating AWS session", err)
	}

	// create an ec2metadata instance
	svc := kb.CreateEC2MetadataService(sess)
	// check for AWs
	if svc.Available() == true {
		log.Info("Running in AWS")
		kb.CloudProvider = awsCloud

		awsHostname, err := svc.GetMetadata(awslocalhostname) // set hostname if in AWS
		splitHostname := strings.Split(awsHostname, ".")
		region, err := svc.Region()

		hostname = splitHostname[0] + "." + region + awsDN
		if len(splitHostname) < 1 {
			log.Warn("Cannot auto detect domainname")
		} else {
			detectedDomainName = splitHostname[1] + "." + splitHostname[2]
		}

		if err != nil {
			log.Fatal("Error getting metadata", err)
		}

		return true
	}

	return false
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kubeadm-bootstrap.yaml)")
	RootCmd.PersistentFlags().StringVarP(&nodeName, "nodename", "n", "", "nodename for bootstrap master")
	RootCmd.PersistentFlags().StringVarP(&datacenter, "datacenter", "d", "", "datacenter name for cluster boostrap")
	RootCmd.PersistentFlags().StringVarP(&clusterName, "clustername", "c", "k1", "cluster name for cluster bootstrap")
	RootCmd.PersistentFlags().StringVarP(&domainName, "domainname", "D", "", "domain name for nodes in cluster")
	RootCmd.PersistentFlags().StringVarP(&kubeadmFile, "kubeadmfile", "f", "/etc/kubernetes/kubeadm.json", "path to kubeadm file to write")
	RootCmd.PersistentFlags().StringVarP(&addressList, "addresslist", "a", "", "comma separated list of IP's for the cluster")
	RootCmd.PersistentFlags().StringVarP(&svcIP, "svcip", "s", "10.96.0.1", "kubernetes service IP")
	RootCmd.PersistentFlags().IntVarP(&numberMasters, "number", "m", 3, "number of masters in the cluster")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "kubernetes bootstrap token")
	RootCmd.PersistentFlags().BoolVarP(&dryrun, "dry-run", "", false, "output the kubeadm config to stdout instead of a file")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".kubeadm-bootstrap") // name of config file (without extension)
	viper.AddConfigPath("$HOME")              // adding home directory as first search path
	viper.AutomaticEnv()                      // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
