# WhatsApp Business API Mockserver

This repository contains the mockserver for the WhatsApp Business API (WABiz).
It is used to perform integration-, system- and load-testing.

## Authentication

The API uses a static API key for authentication. Set the `WA_API_KEY` environment variable with your token.

All requests (except `/generate` and `/generate/cancel`) must include the Authorization header:
```
Authorization: Bearer <your_api_key>
```

## Webhook Configuration

Webhooks are now sent following Meta's official WhatsApp Cloud API format. To configure the webhook metadata, add these fields to your configuration file:

```json
{
  "businessAccountId": "your_business_account_id",
  "phoneNumberId": "your_phone_number_id",
  "displayPhoneNumber": "+1234567890"
}
```

The webhook payload will follow this structure:
```json
{
  "object": "whatsapp_business_account",
  "entry": [{
    "id": "BUSINESS_ACCOUNT_ID",
    "changes": [{
      "field": "messages",
      "value": {
        "messaging_product": "whatsapp",
        "metadata": {
          "display_phone_number": "PHONE_NUMBER",
          "phone_number_id": "PHONE_NUMBER_ID"
        },
        "contacts": [...],
        "messages": [...],
        "statuses": [{
          "id": "wamid.xxx",
          "status": "delivered",
          "recipient_id": "+5511999999999",
          "timestamp": 1734304800,
          "conversation": {
            "id": "conv123"
          },
          "pricing": {
            "pricing_model": "CBP",
            "billable": true
          }
        }]
      }
    }]
  }]
}
```

## Supported Endpoints

| Endpoint| Usage | Supported |
| :--------------- | :------------- | :------ |
| POST /v1/generate| generate webhook requests| ✅ |
| POST /v1/generate/cancel  | stop generation of webhook requests| ✅ |
| POST /v1/{phoneNumberId}/messages| send messages (WhatsApp format)| ✅ |
| POST /v1/media| save media file| ✅ |
| GET /v1/media/{id}| get media file| ✅ |
| DEL /v1/media/{id}| delete media file| ✅ |
| POST /v1/contacts| check for wa_id for contact input| ✅ |
| XXX /v1/contacts/{wa_id}/identity | manage whatsapp id identity| ❌ |
| XXX /v1/settings/**| setup application settings| ✅ |
| XXX /v1/profile/**| setup all profile settings| ✅ |
| XXX /v1/stickerpacks/**| all stickerpacks functionality | ❌ |
| XXX /v1/certificates/**| webhook ca certificates functionality | ✅ |
| XXX /v1/account | registration functionality | ✅ |
| XXX /v1/account/verify | registration functionality | ✅ |
| GET /v1/health | health check endpoint | ✅ |

## Functionality
The following list shows the core functionality that is currently supported.

1. Generate inbound traffic with different messages and media
2. Generate stati for outbound traffic
3. Rate limiting
4. (TBD) Validate outbound traffic
5. (TBD) strict validation (only allow outbound messages to users that have sent a inbound message)

## Supported Messages
The following message types are currently supported.
Inbound types are generated and sent via the webhook.
Outbound types are accepted by the messages resource and validated.

| Type | Inbound | Outbound |
| :--- | :---| :--- |
| Text | ✅ | ✅ |
| Image | ✅ | ✅ |
| Audio | ✅ | ✅ |
| Video | ✅ | ✅ |
| Document | ✅ | ✅ |
| Location | ❌ | ✅ |
| Interactive | ❌ | ✅ |
| Template | ❌ | ✅  |
| Contact (vCard) | ❌ | ✅ |
| Sticker | ❌ | ✅ |
| System | ❌ | ❌ |



## Usage Examples

### Send a text message
```bash
curl -X POST "http://localhost:9090/v1/123456789/messages" \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "+5511999999999",
    "type": "text",
    "text": {
      "body": "Hello World",
      "preview_url": true
    }
  }'
```

### Send a contact (vCard)
```bash
curl -X POST "http://localhost:9090/v1/123456789/messages" \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "+5511999999999",
    "type": "contacts",
    "contacts": [{
      "name": {
        "formatted_name": "João Silva",
        "first_name": "João",
        "last_name": "Silva"
      },
      "phones": [{
        "phone": "+5511988887777",
        "type": "CELL"
      }]
    }]
  }'
```

### Send an image
```bash
curl -X POST "http://localhost:9090/v1/123456789/messages" \
  -H "Authorization: Bearer your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "+5511999999999",
    "type": "image",
    "image": {
      "link": "https://example.com/image.jpg",
      "caption": "Check this out!"
    }
  }'
```

## Notes

### Generate model code
After modifying `.proto` files, regenerate the Go models:
```bash
make build-protoc
```

### Usage locally with http
```bash
WA_API_KEY=local_test WA_WEBHOOK_URL=https://xxx.webhook.cool WA_TLS_DISABLED=1 WA_INSECURE_SKIP_VERIFY=1 go run ./cmd/main.go --configfile=cmd/config.json
```

### Environment Variables
- `WA_API_KEY` - API authentication token (required)
- `WA_WEBHOOK_URL` - Webhook URL for receiving status updates
- `WA_TLS_DISABLED` - Set to `1` to disable TLS
- `WA_INSECURE_SKIP_VERIFY` - Set to `1` to skip webhook certificate validation
- `WA_API_PREFIX` - API path prefix (default: `/v1`)
- `WA_ADDR` - Server address (default: `0.0.0.0:9090`)
- `WA_LOGLEVEL` - Log level (default: `info`)
- `WA_CONFIGFILE` - Path to configuration file
