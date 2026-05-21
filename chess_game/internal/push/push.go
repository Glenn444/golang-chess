package push

import (
	"encoding/json"
	"errors"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

var ErrSubscriptionGone = errors.New("push subscription gone")

func Send(endpoint, p256dh, auth string, p Payload, cfg Config) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	resp, err := webpush.SendNotification(data, &webpush.Subscription{
		Endpoint: endpoint,
		Keys: webpush.Keys{
			P256dh: p256dh,
			Auth:   auth,
		},
	}, &webpush.Options{
		VAPIDPublicKey:  cfg.VAPIDPublicKey,
		VAPIDPrivateKey: cfg.VAPIDPrivateKey,
		Subscriber:      cfg.VAPIDSubject,
		TTL:             60,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 410 {
		return ErrSubscriptionGone
	}
	return nil
}
