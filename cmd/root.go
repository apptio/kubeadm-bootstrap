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

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
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
var dcName string
var token string
var generatedToken string
var dryrun bool
var quiet bool

// Version string
var Version string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kubeadm-bootstrap",
	Short: "Bootstrap a kubernetes cluster using known good config",
	Long: `Generate a kubeadm config for a kubernetes cluster using CIS compatible configuration
using jsonnet templates for the config file`,
	Run: func(cmd *cobra.Command, args []string) {

        if quiet {
            // set logging to /dev/null
            file, err := os.OpenFile("/dev/null", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
            if err == nil {
                log.SetOutput(file)
            } else {
                log.Fatal("Unable to open /dev/null.")
            }
        } else {
            // set logging to stderr
            log.SetOutput(os.Stderr)
        }

		// read static assets
		templateBox, err := rice.FindBox("../lib")
		if err != nil {
			log.Fatal(err)
		}

		tmpl, err := templateBox.String("kubeadm.libsonnet")
		if err != nil {
			log.Fatal(err)
		}

		// create a jsonnet vm
		vm := jsonnet.MakeVM()

		// check for default required vars
		if datacenter == "" {
			log.Info("Auto detecting dc name")
			out, err := exec.Command("facter", "-p", "datacenter").Output()
			if err != nil {
				log.Fatal("Error detecting datacenter from facter: ", err)
			}

			dcName = string(out)
			dcName = strings.TrimSuffix(dcName, "\n")

			if dcName == "" {
				log.Fatal("No datacenter provided")
			}
			log.Info("Datacenter name is: ", dcName)
		} else {
			dcName = datacenter
		}

		if clusterName == "" {
			log.Fatal("Please specify a cluster name")
		}

		// determine if we're in AWS:

		sess, err := session.NewSession()

		if err != nil {
			log.Fatal("Error creating AWS session", err)
		}

		// create an ec2metadata instance
		svc := ec2metadata.New(sess)

		// check for AWs
		if svc.Available() == true {
			log.Info("Running in AWS")
			cloudProvider = "aws"

			awsHostname, err := svc.GetMetadata("local-hostname") // set hostname if in AWS
			splitHostname := strings.Split(awsHostname, ".")
			region, err := svc.Region()

			hostname = splitHostname[0] + "." + region + ".compute.internal"
			if len(splitHostname) < 1 {
				log.Warn("Cannot auto detect domainname")
			} else {
				detectedDomainName = splitHostname[1] + "." + splitHostname[2]
			}

			if err != nil {
				log.Fatal("Error getting metadata", err)
			}

		} else {
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
		}

		if nodeName == "" {
			if hostname == "" {
				log.Fatal("Unable to detect hostname and no hostname provided")
			}
			log.Info("No hostname provided - auto detecting hostname")
			nodeName = hostname
		}

		if domainName == "" {
			if detectedDomainName == "" {
				log.Fatal("Please specify a domain name for the cluster")
			}
			domainName = detectedDomainName
		}

		if addressList == "" {
			addresses = n.GetMasterAddresses(dcName, clusterName, domainName, numberMasters, svcIP)
			//log.Fatal("Please specify an list of IP addresses for the cluster")
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

		// populate jsonnet extvars
		vm.ExtVar("datacenter", dcName)
		vm.ExtVar("clustername", clusterName)
		vm.ExtVar("domainname", domainName)
		vm.ExtVar("nodename", nodeName)
		vm.ExtVar("cloudprovider", cloudProvider)
		vm.ExtVar("ipaddress", ipAddress)
		vm.ExtVar("addresslist", addresses)
		vm.ExtVar("token", token)
		vm.ExtVar("number_masters", strconv.Itoa(numberMasters))

		// evaluate jsonnet snippet
		out, err := vm.EvaluateSnippet("file", tmpl)

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
	RootCmd.PersistentFlags().BoolVarP(&quiet, "quiet" , "", false, "suppress logging output")

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
