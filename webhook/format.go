package webhook

import (
	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

// WhatsAppWebhookPayload represents the complete webhook payload following Meta's format
type WhatsAppWebhookPayload struct {
	Object string                   `json:"object"`
	Entry  []WhatsAppWebhookEntry   `json:"entry"`
}

// WhatsAppWebhookEntry represents an entry in the webhook payload
type WhatsAppWebhookEntry struct {
	ID      string                    `json:"id"`
	Changes []WhatsAppWebhookChange   `json:"changes"`
}

// WhatsAppWebhookChange represents a change in a webhook entry
type WhatsAppWebhookChange struct {
	Field string                 `json:"field"`
	Value WhatsAppWebhookValue   `json:"value"`
}

// WhatsAppWebhookValue represents the value object in a webhook change
type WhatsAppWebhookValue struct {
	MessagingProduct string                    `json:"messaging_product"`
	Metadata         WhatsAppWebhookMetadata   `json:"metadata"`
	Contacts         []*model.Contact          `json:"contacts,omitempty"`
	Messages         []*model.Message          `json:"messages,omitempty"`
	Statuses         []*WebhookStatus          `json:"statuses,omitempty"`
	Errors           []*model.Error            `json:"errors,omitempty"`
}

// WhatsAppWebhookMetadata represents the metadata for a webhook notification
type WhatsAppWebhookMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

// WebhookStatus represents a status in the webhook with string status field
type WebhookStatus struct {
	ID           string              `json:"id"`
	Status       string              `json:"status"`
	RecipientID  string              `json:"recipient_id,omitempty"`
	Timestamp    int64               `json:"timestamp,omitempty"`
	Conversation *model.Conversation `json:"conversation,omitempty"`
	Pricing      *WebhookPricing     `json:"pricing,omitempty"`
}

// WebhookPricing represents pricing in the webhook with string pricing_model field
type WebhookPricing struct {
	PricingModel string `json:"pricing_model"`
	Billable     bool   `json:"billable"`
}

// statusEnumToString converts StatusEnum to its string representation
func statusEnumToString(status model.Status_StatusEnum) string {
	switch status {
	case model.Status_sent:
		return "sent"
	case model.Status_delivered:
		return "delivered"
	case model.Status_read:
		return "read"
	default:
		return "unknown"
	}
}

// pricingModelToString converts PricingModel enum to its string representation
func pricingModelToString(pricingModel model.Pricing_PricingModel) string {
	switch pricingModel {
	case model.Pricing_CBP:
		return "CBP"
	case model.Pricing_NBP:
		return "NBP"
	default:
		return "unknown"
	}
}

// ConvertToMetaFormat converts a WebhookRequest to the Meta webhook format
func ConvertToMetaFormat(req *model.WebhookRequest, businessAccountID, phoneNumberID, displayPhoneNumber string) *WhatsAppWebhookPayload {
	// Create the metadata
	metadata := WhatsAppWebhookMetadata{
		DisplayPhoneNumber: displayPhoneNumber,
		PhoneNumberID:      phoneNumberID,
	}

	// Convert statuses to webhook format with string status field
	webhookStatuses := make([]*WebhookStatus, len(req.Statuses))
	for i, status := range req.Statuses {
		var webhookPricing *WebhookPricing
		if status.Pricing != nil {
			webhookPricing = &WebhookPricing{
				PricingModel: pricingModelToString(status.Pricing.PricingModel),
				Billable:     status.Pricing.Billable,
			}
		}

		webhookStatuses[i] = &WebhookStatus{
			ID:           status.Id,
			Status:       statusEnumToString(status.Status),
			RecipientID:  status.RecipientId,
			Timestamp:    status.Timestamp,
			Conversation: status.Conversation,
			Pricing:      webhookPricing,
		}
	}

	// Create the value object
	value := WhatsAppWebhookValue{
		MessagingProduct: "whatsapp",
		Metadata:         metadata,
		Contacts:         req.Contacts,
		Messages:         req.Messages,
		Statuses:         webhookStatuses,
		Errors:           req.Errors,
	}

	// Create the change object
	change := WhatsAppWebhookChange{
		Field: "messages",
		Value: value,
	}

	// Create the entry object
	entry := WhatsAppWebhookEntry{
		ID:      businessAccountID,
		Changes: []WhatsAppWebhookChange{change},
	}

	// Create the complete payload
	payload := &WhatsAppWebhookPayload{
		Object: "whatsapp_business_account",
		Entry:  []WhatsAppWebhookEntry{entry},
	}

	return payload
}