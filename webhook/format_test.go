package webhook

import (
	"encoding/json"
	"testing"

	"github.com/ron96G/whatsapp-bizapi-mock/model"
)

func TestStatusEnumToString(t *testing.T) {
	tests := []struct {
		name     string
		status   model.Status_StatusEnum
		expected string
	}{
		{"sent status", model.Status_sent, "sent"},
		{"delivered status", model.Status_delivered, "delivered"},
		{"read status", model.Status_read, "read"},
		{"unknown status", model.Status_unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statusEnumToString(tt.status)
			if result != tt.expected {
				t.Errorf("statusEnumToString(%v) = %v, want %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestPricingModelToString(t *testing.T) {
	tests := []struct {
		name     string
		pricing  model.Pricing_PricingModel
		expected string
	}{
		{"CBP model", model.Pricing_CBP, "CBP"},
		{"NBP model", model.Pricing_NBP, "NBP"},
		{"unknown model", model.Pricing_unknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pricingModelToString(tt.pricing)
			if result != tt.expected {
				t.Errorf("pricingModelToString(%v) = %v, want %v", tt.pricing, result, tt.expected)
			}
		})
	}
}

func TestConvertToMetaFormat_StatusAsString(t *testing.T) {
	// Create a test WebhookRequest with a status
	req := &model.WebhookRequest{
		Statuses: []*model.Status{
			{
				Id:          "wamid.test123",
				Status:      model.Status_delivered,
				RecipientId: "+5511999999999",
				Timestamp:   1734304800,
				Conversation: &model.Conversation{
					Id: "conv123",
				},
				Pricing: &model.Pricing{
					PricingModel: model.Pricing_CBP,
					Billable:     true,
				},
			},
		},
	}

	// Convert to Meta format
	payload := ConvertToMetaFormat(req, "business123", "phone456", "+1234567890")

	// Marshal to JSON to verify the format
	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Verify the status is a string in JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Navigate to the status field
	entry := result["entry"].([]interface{})[0].(map[string]interface{})
	changes := entry["changes"].([]interface{})[0].(map[string]interface{})
	value := changes["value"].(map[string]interface{})
	statuses := value["statuses"].([]interface{})
	status := statuses[0].(map[string]interface{})

	// Check that status field is a string, not a number
	statusValue, ok := status["status"].(string)
	if !ok {
		t.Errorf("Status field is not a string, got type: %T", status["status"])
	}

	if statusValue != "delivered" {
		t.Errorf("Status value = %v, want 'delivered'", statusValue)
	}

	// Check pricing_model is also a string
	pricing := status["pricing"].(map[string]interface{})
	pricingModel, ok := pricing["pricing_model"].(string)
	if !ok {
		t.Errorf("Pricing model field is not a string, got type: %T", pricing["pricing_model"])
	}

	if pricingModel != "CBP" {
		t.Errorf("Pricing model = %v, want 'CBP'", pricingModel)
	}

	t.Logf("Generated webhook JSON:\n%s", string(jsonData))
}