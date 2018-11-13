package notifiers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
)

/*
The Slack Notifier takes in an EC2 Instance Status and will craft a message and send it to Slack via a webhook.
*/

// Slack holds the Slack Webhook URL
type Slack struct {
	WebhookUR string
}

// Field holds field level values for the Slack message
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Attachment contains the Slack Message Attachment fields
type Attachment struct {
	Fallback string   `json:"fallback"`
	Color    string   `json:"color"`
	Fields   []*Field `json:"fields"`
}

// Payload is the message payload to be sent to Slack
type Payload struct {
	Username    string        `json:"username,omitempty"`
	Text        string        `json:"text,omitempty"`
	Attachments []*Attachment `json:"attachments,omitempty"`
}

// AddField will add the field to the Attachment
func (attachment *Attachment) AddField(field *Field) *Attachment {
	attachment.Fields = append(attachment.Fields, field)
	return attachment
}

// AddAttachment adds the Attachment to the Payload
func (payload *Payload) AddAttachment(attachment *Attachment) *Payload {
	payload.Attachments = append(payload.Attachments, attachment)
	return payload
}

// Regex used to check if an event is Completed and should not be sent in the notification
var completedRegexp = regexp.MustCompile("^Completed")

// Notify will craft a Slack Payload and post the message to the Webhook URL.
func (s *Slack) Notify(instance *ec2.InstanceStatus) error {
	// Create our Slack message payload
	payload := &Payload{
		Attachments: make([]*Attachment, 0),
	}
	payload.Text = "There are scheduled EC2 Events"
	payload.Username = "EC2 Event Notifier"
	for _, event := range instance.Events {
		// If the event description is "Completed", don't report the event
		if completedRegexp.MatchString(*event.Description) {
			continue
		}
		log.Println("Found EC2 Event, notifying Slack")

		attachment := &Attachment{}

		if event.NotBefore.Before(time.Now()) {
			attachment.Color = "danger"
		} else {
			attachment.Color = "warning"
		}

		attachment.Fallback = fmt.Sprintf("%s / %s / %s - %s / %s",
			*instance.InstanceId, *event.Code, event.NotBefore, event.NotAfter, *event.Description)

		// Instance ID Field
		attachment.AddField(&Field{
			Title: "Instance",
			Value: *instance.InstanceId,
			Short: true,
		})

		// Event Code Field
		attachment.AddField(&Field{
			Title: "Event Type",
			Value: *event.Code,
			Short: true,
		})

		// Duration Field
		attachment.AddField(&Field{
			Title: "Duration",
			Value: fmt.Sprintf("%s - %s", event.NotBefore, event.NotAfter),
			Short: false,
		})

		// Description Field
		attachment.AddField(&Field{
			Title: "Description",
			Value: *event.Description,
			Short: false,
		})

		payload.AddAttachment(attachment)
	}
	// If we added any attachments, send the message to Slack.
	if len(payload.Attachments) > 0 {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		buf := bytes.NewReader(data)
		if _, err := http.DefaultClient.Post(s.WebhookUR, "application/json", buf); err != nil {
			return err
		}
	}
	return nil
}
