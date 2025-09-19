package dto

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

func (n Settings) Value() (driver.Value, error) { return valueJSON(n) }
func (n *Settings) Scan(value any) error        { return scanJSON(n, value) }

func (n MetadataMap) Value() (driver.Value, error) { return valueJSON(n) }
func (n *MetadataMap) Scan(value any) error        { return scanJSON(n, value) }

func (n WebhookResponse) Value() (driver.Value, error) { return valueJSON(n) }
func (n *WebhookResponse) Scan(value any) error        { return scanJSON(n, value) }

func (n WebhookRequest) Value() (driver.Value, error) { return valueJSON(n) }
func (n *WebhookRequest) Scan(value any) error        { return scanJSON(n, value) }

func (n WebhookExecution) Value() (driver.Value, error) { return valueJSON(n) }
func (n *WebhookExecution) Scan(value any) error        { return scanJSON(n, value) }

func (n Webhook) Value() (driver.Value, error) { return valueJSON(n) }
func (n *Webhook) Scan(value any) error        { return scanJSON(n, value) }

func (n DirectWebhooks) Value() (driver.Value, error) { return valueJSON(n) }
func (n *DirectWebhooks) Scan(value any) error        { return scanJSON(n, value) }

func (n NewWebhook) Value() (driver.Value, error) { return valueJSON(n) }
func (n *NewWebhook) Scan(value any) error        { return scanJSON(n, value) }

func (p PrePostProcessing) Value() (driver.Value, error) { return valueJSON(p) }
func (p *PrePostProcessing) Scan(value any) error        { return scanJSON(p, value) }

func (p RawResolved) Value() (driver.Value, error) { return valueJSON(p) }
func (p *RawResolved) Scan(value any) error        { return scanJSON(p, value) }

func (n NewPrePostProcessing) Value() (driver.Value, error) { return valueJSON(n) }
func (n *NewPrePostProcessing) Scan(value any) error        { return scanJSON(n, value) }

func scanJSON[T any](dst *T, value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed (got %T)", value)
	}
	return json.Unmarshal(bytes, dst)
}

func valueJSON[T any](src T) (driver.Value, error) {
	return json.Marshal(src)
}
