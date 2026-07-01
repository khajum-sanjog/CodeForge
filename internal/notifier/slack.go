package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SlackNotifier sends a formatted webhook request using Slack Block Kit.
type SlackNotifier struct {
	WebhookURL string
}

// NewSlackNotifier returns a new SlackNotifier instance.
func NewSlackNotifier(webhookURL string) *SlackNotifier {
	return &SlackNotifier{WebhookURL: webhookURL}
}

// Send posts a formatted deployment update to the Slack webhook.
func (s *SlackNotifier) Send(payload Payload) error {
	if s.WebhookURL == "" {
		return fmt.Errorf("Slack Webhook URL is empty")
	}

	color := "#534AB7" // CodeForge primary color (purple)
	emoji := "●"
	statusStr := strings.ToUpper(payload.Status)
	switch strings.ToLower(payload.Status) {
	case "success":
		color = "#0F6E56" // success color (green)
		emoji = "✓"
	case "failed", "error":
		color = "#D85A30" // error color (red)
		emoji = "✗"
	case "rollback":
		color = "#BA7517" // warning color (amber)
		emoji = "↺"
	}

	// Build block kit structure
	slackPayload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"blocks": []map[string]interface{}{
					{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": fmt.Sprintf("*CodeForge Deployment* %s *%s*", emoji, statusStr),
						},
					},
					{
						"type": "section",
						"fields": []map[string]interface{}{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Project:*\n%s", payload.Project),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Trigger:*\n%s", payload.Trigger),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Duration:*\n%s", payload.Duration.Round(time.Millisecond).String()),
							},
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("*Commit SHA:*\n%s", formatCommit(payload.CommitSHA)),
							},
						},
					},
				},
			},
		},
	}

	if payload.ErrorMsg != "" {
		// Append error block
		attach := slackPayload["attachments"].([]map[string]interface{})[0]
		blocks := attach["blocks"].([]map[string]interface{})
		blocks = append(blocks, map[string]interface{}{
			"type": "section",
			"text": map[string]string{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*Error details:*\n`%s`", payload.ErrorMsg),
			},
		})
		attach["blocks"] = blocks
	}

	data, err := json.Marshal(slackPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.WebhookURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Slack webhook failed with status code %d", resp.StatusCode)
	}

	return nil
}

func formatCommit(sha string) string {
	if sha == "" {
		return "N/A"
	}
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
