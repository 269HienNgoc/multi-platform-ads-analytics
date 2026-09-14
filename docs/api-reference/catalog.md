# Advertising catalog API

The catalog API writes and reads the provider-neutral hierarchy:

`Platform -> Ad Account -> Campaign -> Ad Group/Ad Set -> Ad -> Creative`

All create endpoints return `201 Created`. Child requests use the internal UUID returned by their parent endpoint; provider IDs belong in `external_id` and are never database primary keys.

## Routes

| Method | Route | Purpose |
| --- | --- | --- |
| POST | `/api/v1/ad-accounts` | Create a provider account |
| POST | `/api/v1/campaigns` | Create a campaign under an account |
| POST | `/api/v1/ad-groups` | Create an ad group/ad set under a campaign |
| POST | `/api/v1/ads` | Create an ad under an ad group |
| POST | `/api/v1/creatives` | Create a creative under an ad |
| GET | `/api/v1/ad-accounts/{accountID}/hierarchy` | Read the complete normalized subtree |

Every create request accepts a `provider_data` JSON object for fields that do not belong in the normalized cross-platform model. Duplicate external IDs under the same parent return `409 Conflict`; missing parents return `404 Not Found`; malformed input returns `400 Bad Request`.

Authentication and authorization are intentionally not exposed yet. Do not publish the API directly to the internet until that boundary is implemented.
