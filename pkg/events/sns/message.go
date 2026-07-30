// Package sns parses and verifies AWS SNS HTTP(S) deliveries: message
// envelope validation, signing certificate retrieval, RSA signature
// verification, and subscription URL validation.
package sns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"regexp"
	"strings"
)

var (
	regionPattern = regexp.MustCompile(`^[a-z]{2,4}(?:-[a-z0-9]+)+-\d+$`)
	certPattern   = regexp.MustCompile(`^SimpleNotificationService-[[:xdigit:]]+\.pem$`)
)

// Message is an SNS delivery: a Notification, SubscriptionConfirmation, or
// UnsubscribeConfirmation.
type Message struct {
	Type             string `json:"Type"`
	MessageID        string `json:"MessageId"`
	TopicARN         string `json:"TopicArn"`
	Subject          string `json:"Subject,omitempty"`
	Message          string `json:"Message"`
	Timestamp        string `json:"Timestamp"`
	SignatureVersion string `json:"SignatureVersion"`
	Signature        string `json:"Signature"`
	SigningCertURL   string `json:"SigningCertURL"`
	Token            string `json:"Token,omitempty"`
	SubscribeURL     string `json:"SubscribeURL,omitempty"`
}

// ParseMessage decodes and validates an SNS message body, including its
// SigningCertURL and, for confirmations, its SubscribeURL.
func ParseMessage(body []byte) (*Message, error) {
	var msg Message
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&msg); err != nil {
		return nil, fmt.Errorf("decode SNS message: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("invalid trailing SNS message data")
	}
	if msg.Type == "" || msg.MessageID == "" || msg.TopicARN == "" || msg.Message == "" || msg.Timestamp == "" || msg.SignatureVersion == "" || msg.Signature == "" || msg.SigningCertURL == "" {
		return nil, errors.New("SNS message is missing required fields")
	}
	switch msg.Type {
	case "Notification":
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		if msg.Token == "" || msg.SubscribeURL == "" {
			return nil, errors.New("SNS confirmation is missing required fields")
		}
		if err := ValidateSubscribeURL(msg.SubscribeURL, msg.TopicARN, msg.Token); err != nil {
			return nil, fmt.Errorf("invalid SNS SubscribeURL: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported SNS message type %q", msg.Type)
	}
	if msg.SignatureVersion != "1" && msg.SignatureVersion != "2" {
		return nil, fmt.Errorf("unsupported SNS signature version %q", msg.SignatureVersion)
	}
	if err := validateURL(msg.SigningCertURL, true); err != nil {
		return nil, fmt.Errorf("invalid SNS SigningCertURL: %w", err)
	}
	return &msg, nil
}

func (m *Message) canonicalString() (string, error) {
	var fields []string
	switch m.Type {
	case "Notification":
		fields = []string{"Message", m.Message, "MessageId", m.MessageID}
		if m.Subject != "" {
			fields = append(fields, "Subject", m.Subject)
		}
		fields = append(fields, "Timestamp", m.Timestamp, "TopicArn", m.TopicARN, "Type", m.Type)
	case "SubscriptionConfirmation", "UnsubscribeConfirmation":
		fields = []string{"Message", m.Message, "MessageId", m.MessageID, "SubscribeURL", m.SubscribeURL, "Timestamp", m.Timestamp, "Token", m.Token, "TopicArn", m.TopicARN, "Type", m.Type}
	default:
		return "", fmt.Errorf("unsupported SNS message type %q", m.Type)
	}
	return strings.Join(fields, "\n") + "\n", nil
}

func validateURL(rawURL string, certificate bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Port() != "" {
		return errors.New("SNS URL must use unadorned HTTPS on the default port")
	}
	host := strings.ToLower(u.Hostname())
	suffix := ".amazonaws.com"
	if strings.HasSuffix(host, ".amazonaws.com.cn") {
		suffix = ".amazonaws.com.cn"
	}
	if !strings.HasPrefix(host, "sns.") || !strings.HasSuffix(host, suffix) {
		return errors.New("SNS URL has an untrusted host")
	}
	region := strings.TrimSuffix(strings.TrimPrefix(host, "sns."), suffix)
	if !regionPattern.MatchString(region) {
		return errors.New("SNS URL has an invalid region")
	}
	if certificate && (!certPattern.MatchString(path.Base(u.EscapedPath())) || strings.Contains(strings.ToLower(u.EscapedPath()), "%2f")) {
		return errors.New("SNS certificate URL has an invalid path")
	}
	return nil
}

func validateHost(rawURL, expectedHost string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if !strings.EqualFold(u.Hostname(), expectedHost) {
		return errors.New("SNS URL host does not match the topic region")
	}
	return nil
}

func hostForTopicARN(topicARN string) (string, error) {
	parts := strings.Split(topicARN, ":")
	if len(parts) < 6 || parts[0] != "arn" || parts[2] != "sns" || !regionPattern.MatchString(parts[3]) {
		return "", errors.New("invalid SNS topic ARN")
	}
	suffix := "amazonaws.com"
	switch parts[1] {
	case "aws", "aws-us-gov":
	case "aws-cn":
		suffix = "amazonaws.com.cn"
	default:
		return "", errors.New("unsupported SNS topic ARN partition")
	}
	return "sns." + parts[3] + "." + suffix, nil
}

// ValidateSubscribeURL rejects a SubscribeURL unless it is a trusted SNS
// endpoint whose ConfirmSubscription query matches the signed message.
func ValidateSubscribeURL(rawURL, topicARN, token string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	query := u.Query()
	if len(query) != 3 || len(query["Action"]) != 1 || len(query["TopicArn"]) != 1 || len(query["Token"]) != 1 ||
		query.Get("Action") != "ConfirmSubscription" || query.Get("TopicArn") != topicARN || query.Get("Token") != token {
		return errors.New("SNS SubscribeURL query does not match the signed message")
	}
	cleanURL := *u
	cleanURL.RawQuery = ""
	if err := validateURL(cleanURL.String(), false); err != nil {
		return err
	}
	if u.EscapedPath() != "/" {
		return errors.New("SNS SubscribeURL has an invalid path")
	}
	return nil
}
