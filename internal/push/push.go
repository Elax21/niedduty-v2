// Package push verschickt Web-Push-Benachrichtigungen an die abonnierten
// Geräte. Es gibt bewusst keinen externen Dienst: der Server signiert die
// Nachrichten selbst per VAPID und schickt sie direkt an den Push-Dienst des
// jeweiligen Browsers.
//
// Das VAPID-Schlüsselpaar wird beim ersten Start erzeugt und in der Tabelle
// `settings` abgelegt — es muss über Neustarts stabil bleiben, sonst werden
// alle bestehenden Abos ungültig. Alternativ per Umgebungsvariablen
// VAPID_PUBLIC_KEY / VAPID_PRIVATE_KEY vorgeben.
package push

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/alessandro/niedduty/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	settingPublicKey  = "vapid_public_key"
	settingPrivateKey = "vapid_private_key"
	// Kontaktadresse für den Push-Dienst (Pflichtangabe im VAPID-Claim).
	subscriber = "mailto:admin@aramaeer-ahlen.de"
	// Nach so vielen Fehlversuchen gilt ein Abo als tot und wird entfernt.
	maxFailures = 3
)

// Keys — das VAPID-Schlüsselpaar dieses Servers.
type Keys struct {
	Public  string
	Private string
}

var (
	mu     sync.Mutex
	loaded *Keys
)

// LoadKeys liefert das Schlüsselpaar und erzeugt es beim ersten Aufruf.
func LoadKeys(db *gorm.DB) (*Keys, error) {
	mu.Lock()
	defer mu.Unlock()
	if loaded != nil {
		return loaded, nil
	}
	if pub, priv := os.Getenv("VAPID_PUBLIC_KEY"), os.Getenv("VAPID_PRIVATE_KEY"); pub != "" && priv != "" {
		loaded = &Keys{Public: pub, Private: priv}
		return loaded, nil
	}
	pub, priv := setting(db, settingPublicKey), setting(db, settingPrivateKey)
	if pub == "" || priv == "" {
		newPriv, newPub, err := webpush.GenerateVAPIDKeys()
		if err != nil {
			return nil, err
		}
		pub, priv = newPub, newPriv
		db.Save(&models.Setting{Key: settingPublicKey, Value: pub})
		db.Save(&models.Setting{Key: settingPrivateKey, Value: priv})
		log.Printf("Push: neues VAPID-Schlüsselpaar erzeugt")
	}
	loaded = &Keys{Public: pub, Private: priv}
	return loaded, nil
}

func setting(db *gorm.DB, key string) string {
	var s models.Setting
	if err := db.First(&s, "key = ?", key).Error; err != nil {
		return ""
	}
	return s.Value
}

// Payload — Inhalt einer Benachrichtigung (wird im Service Worker gelesen).
type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

// SendToUsers schickt eine Nachricht an alle Geräte der angegebenen Konten.
// Rückgabe: Anzahl erfolgreich zugestellter Geräte.
func SendToUsers(db *gorm.DB, userIDs []uuid.UUID, p Payload) int {
	if len(userIDs) == 0 {
		return 0
	}
	var subs []models.PushSubscription
	db.Where("user_id IN ?", userIDs).Find(&subs)
	return send(db, subs, p)
}

// SendToAll schickt eine Nachricht an jedes abonnierte Gerät.
func SendToAll(db *gorm.DB, p Payload) int {
	var subs []models.PushSubscription
	db.Find(&subs)
	return send(db, subs, p)
}

func send(db *gorm.DB, subs []models.PushSubscription, p Payload) int {
	if len(subs) == 0 {
		return 0
	}
	keys, err := LoadKeys(db)
	if err != nil {
		log.Printf("Push: Schlüssel nicht verfügbar: %v", err)
		return 0
	}
	body, err := json.Marshal(p)
	if err != nil {
		return 0
	}
	sent := 0
	for _, s := range subs {
		sub := &webpush.Subscription{
			Endpoint: s.Endpoint,
			Keys:     webpush.Keys{P256dh: s.P256dh, Auth: s.Auth},
		}
		resp, err := webpush.SendNotification(body, sub, &webpush.Options{
			Subscriber:      subscriber,
			VAPIDPublicKey:  keys.Public,
			VAPIDPrivateKey: keys.Private,
			TTL:             60 * 60 * 6,
			Urgency:         webpush.UrgencyNormal,
		})
		if err != nil {
			markFailed(db, s, err.Error())
			continue
		}
		resp.Body.Close()
		switch {
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
			// Abo vom Browser gelöscht — endgültig entfernen.
			db.Delete(&models.PushSubscription{}, "id = ?", s.ID)
		case resp.StatusCode >= 400:
			markFailed(db, s, resp.Status)
		default:
			sent++
			if s.Failures > 0 {
				db.Model(&models.PushSubscription{}).Where("id = ?", s.ID).Update("failures", 0)
			}
		}
	}
	return sent
}

func markFailed(db *gorm.DB, s models.PushSubscription, reason string) {
	s.Failures++
	if s.Failures >= maxFailures {
		db.Delete(&models.PushSubscription{}, "id = ?", s.ID)
		log.Printf("Push: Abo entfernt nach %d Fehlversuchen (%s)", s.Failures, reason)
		return
	}
	db.Model(&models.PushSubscription{}).Where("id = ?", s.ID).Update("failures", s.Failures)
}
