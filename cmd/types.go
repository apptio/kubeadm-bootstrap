package cmd

import (
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Bootstrap interface {
	getRiceInformation(string, string) (string, error)
	getDCInformation() string
	getAWSInformation(string)
	getAWSSession(string) (*session.Session, error)
	getEC2Metadata() *ec2metadata.EC2Metadata
}

type KubeBootstrap struct {
	DCName        string
	Cluster       string
	DomainName    string
	NodeName      string
	CloudProvider string
	IPAddress     string
	Addresses     string
	Token         string
	NumMasters    string
}

func NewKubeBootstrap(dc string, cn string) *KubeBootstrap {
	return &KubeBootstrap{
		DCName:  dc,
		Cluster: cn,
	}
}

func (kb *KubeBootstrap) GetAWSClusterInformation() {

}

func (kb *KubeBootstrap) CreateAWSSession(string) (*session.Session, error) {
	return nil, nil
}

func (kb *KubeBootstrap) CreateEC2MetadataService(session *session.Session) *ec2metadata.EC2Metadata {
	// create an ec2metadata instance
	svc := ec2metadata.New(session)
	return svc
}
