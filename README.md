# EC2 Event Notifier

The EC2 Event Notifier is a simple utility that can send notifications for EC2 Events.  An EC2 Event is a scheduled event
by AWS, usually for alerting that an instance needs to be restarted or retired.  This utility will look at all AWS EC2
instances and send a notification when an event is found.

The EC2 Event Notifier currently supports Slack Webhooks.  To enable the Slack Webhook, start the application with 
the `-slackWebHookURL` flag.

Example:
```bash
ec2EventNotifyer -slackWebHookURL https://hooks.slack.com/services/...
```

The EC2 Event Notifier uses the native [AWS SDK for Go](https://github.com/aws/aws-sdk-go).  As such, it will look for
either environment variables or the `.config` file setup with `aws config`.

## Adding Additional Notifiers

You can add additional Notifiers, such as an E-Mail notifier, Pager Duty, etc.  To create a new Notifier, it must
implement the `Notifier` interface.

```golang
type Notifier interface {
    Notify(*ec2.InstanceStatus) error
}
```

The Notifier Interface has a single function `Notify` that takes in an EC2 Instance Status pointer.  The EC2 Instance
Status object holds the instance information as well as a slice of Events.  The `Notify` function is responsible for
creating the notification message and sending the notification.

To register a new Notifier, add an instance of it to the `toNotify` slice.

```golang
toNotify := make([]notifiers.Notifier, 0)

// Setup Slack Notifier
if slackWebHookURL != "" {
	log.Println("Enabling Slack Notifier")
	toNotify = append(toNotify, &notifiers.Slack{WebhookUR: slackWebHookURL})
}
``` 

Please see the Slack implementation found in `notifiers/slack.go` for an example of creating a Notifier.