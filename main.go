package main

import (
	"flag"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/kenXengineering/ec2EventNotifier/notifiers"
)

var (
	// Slack Configuration
	slackWebHookURL string
)

func main() {
	// Slack Webhook URL Flag
	flag.StringVar(&slackWebHookURL, "slackWebHookURL", "", "Slack WebHook URL.  Setting this field enables notifications to Slack.")
	flag.Parse()

	// List of Notifiers to call when an EC2 Event is found
	toNotify := make([]notifiers.Notifier, 0)

	// Setup Slack Notifier
	if slackWebHookURL != "" {
		log.Println("Enabling Slack Notifier")
		toNotify = append(toNotify, &notifiers.Slack{WebhookUR: slackWebHookURL})
	}

	// Check if any notifiers are enabled
	if len(toNotify) == 0 {
		log.Println("No notifiers setup, exiting")
		os.Exit(0)
	}

	// Create our AWS Session and EC2 Client
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		log.Fatal(err)
	}
	ec2Svc := ec2.New(sess)

	// Check for EC2 Events
	log.Println("Checking for EC2 Events")
	CheckForEC2Events(ec2Svc, toNotify)
}

// CheckForEC2Events takes in the EC2 Service and a list of Notifiers to notify when an EC2 Event is found.  The function
// will get a list of EC2 Instances and Statuses.  If an instance has events, it will pass the instance status to the
// notifier
func CheckForEC2Events(ec2Svc *ec2.EC2, notifiers []notifiers.Notifier) {
	// Get our instances status
	instances, err := GetEC2InstanceStatus(ec2Svc)
	if err != nil {
		log.Fatal(err)
	}
	// Loop over the instances and see if any of them have any events
	for _, instance := range instances {
		if len(instance.Events) > 0 {
			// We found an instance with events, notify all the notifiers
			for _, notifier := range notifiers {
				if err := notifier.Notify(instance); err != nil {
					log.Println(err)
				}
			}

		}
	}
}

// GetEc2InstanceStatus returns a list of EC2 InstanceStatus objects.  It will describe the status for all EC2 instances,
// including stopped instances.
func GetEC2InstanceStatus(ec2Svc *ec2.EC2) ([]*ec2.InstanceStatus, error) {
	instances := make([]*ec2.InstanceStatus, 0)
	// Get events for all instances, even stopped ones.
	input := &ec2.DescribeInstanceStatusInput{
		IncludeAllInstances: aws.Bool(true),
	}
	// Loop over response.  If there are more then 1000 instances we must make another call using the nextToken value
	nextToken := ""
	for {
		if nextToken != "" {
			input.NextToken = &nextToken
		}
		output, err := ec2Svc.DescribeInstanceStatus(input)
		if err != nil {
			return nil, err
		}
		instances = append(instances, output.InstanceStatuses...)
		// If we found a token, then there are more instances to request
		if output.NextToken != nil {
			nextToken = *output.NextToken
			continue
		}
		break
	}

	return instances, nil
}
