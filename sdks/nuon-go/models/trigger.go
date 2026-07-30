package models

import "time"

type EventFieldSelector struct {
	Header  string `json:"header,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type TriggerAuthConfig struct {
	Header          string   `json:"header,omitempty"`
	Prefix          string   `json:"prefix,omitempty"`
	Encoding        string   `json:"encoding,omitempty"`
	Algorithm       string   `json:"algorithm,omitempty"`
	Username        string   `json:"username,omitempty"`
	Issuer          string   `json:"issuer,omitempty"`
	Audience        []string `json:"audience,omitempty"`
	TopicARN        string   `json:"topic_arn,omitempty"`
	ExpectedEmail   string   `json:"expected_email,omitempty"`
	ExpectedSubject string   `json:"expected_subject,omitempty"`
}

type TriggerCreateRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Preset      string             `json:"preset,omitempty"`
	Secret      string             `json:"secret,omitempty"`
	AuthType    string             `json:"auth_type,omitempty"`
	AuthConfig  TriggerAuthConfig  `json:"auth_config,omitempty"`
	Envelope    string             `json:"envelope,omitempty"`
	TypeFrom    EventFieldSelector `json:"type_from,omitempty"`
	IDFrom      EventFieldSelector `json:"id_from,omitempty"`
}

type TriggerSecret struct {
	ID         string     `json:"id"`
	KeyID      string     `json:"key_id"`
	CreatedAt  time.Time  `json:"created_at"`
	NotBefore  time.Time  `json:"not_before"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type Trigger struct {
	ID          string             `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	AuthType    string             `json:"auth_type"`
	Preset      string             `json:"preset,omitempty"`
	AuthConfig  TriggerAuthConfig  `json:"auth_config,omitempty"`
	Envelope    string             `json:"envelope"`
	TypeFrom    EventFieldSelector `json:"type_from,omitempty"`
	IDFrom      EventFieldSelector `json:"id_from,omitempty"`
	Status      string             `json:"status"`
	LastEventAt *time.Time         `json:"last_event_at,omitempty"`
	Secrets     []TriggerSecret    `json:"secrets,omitempty"`
}

type TriggerCredentialResponse struct {
	Trigger    Trigger `json:"trigger"`
	IngressURL string  `json:"ingress_url,omitempty"`
	KeyID      string  `json:"key_id,omitempty"`
	Secret     string  `json:"secret,omitempty"`
}

type TriggerRevokeResponse struct {
	RevokedAt time.Time `json:"revoked_at"`
}

type TriggerIngressURLResponse struct {
	IngressURL string `json:"ingress_url"`
}

type TriggerSecretRevealResponse struct {
	KeyID  string `json:"key_id"`
	Secret string `json:"secret"`
}
