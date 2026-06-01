package notifier

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neuralops/platform/internal/alerting/model"
	"github.com/neuralops/platform/internal/domain"
	"github.com/neuralops/platform/internal/mobile"
)

// MobilePushNotifier sends alerts to registered Expo/FCM devices.
type MobilePushNotifier struct {
	expo *mobile.ExpoClient
}

// NewMobilePushNotifier creates a mobile push notifier.
func NewMobilePushNotifier(pool *pgxpool.Pool) *MobilePushNotifier {
	return &MobilePushNotifier{expo: mobile.NewExpoClient(pool)}
}

func (n *MobilePushNotifier) Type() string { return "mobile_push" }

func (n *MobilePushNotifier) Send(ctx context.Context, alert model.AlertRecord, _ *domain.Incident) error {
	if n.expo == nil {
		return fmt.Errorf("mobile push unavailable")
	}
	title := alert.Title
	if title == "" {
		title = alert.AlertName
	}
	body := fmt.Sprintf("[%s] %s — %s", alert.Severity, alert.Service, alert.Description)
	return n.expo.SendToTenant(ctx, alert.TenantID, title, body)
}
