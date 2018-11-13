package notifiers

import "github.com/aws/aws-sdk-go/service/ec2"

// Notifier interface
type Notifier interface {
	Notify(*ec2.InstanceStatus) error
}
