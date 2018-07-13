package notifiers

import (
	"fmt"
	"github.com/hunterlong/statup/utils"
	"strings"
	"time"
	"upper.io/db.v3"
)

var (
	AllCommunications []AllNotifiers
	Collections       db.Collection
)

type AllNotifiers interface{}

func add(c interface{}) {
	AllCommunications = append(AllCommunications, c)
}

func Load() []AllNotifiers {
	utils.Log(1, "Loading notifiers")
	var notifiers []AllNotifiers
	for _, comm := range AllCommunications {
		n := comm.(Notifier)
		n.Init()
		notifiers = append(notifiers, n)
		n.Test()
	}
	return notifiers
}

type Notification struct {
	Id        int64     `db:"id,omitempty" json:"id"`
	Method    string    `db:"method" json:"method"`
	Host      string    `db:"host" json:"-"`
	Port      int       `db:"port" json:"-"`
	Username  string    `db:"username" json:"-"`
	Password  string    `db:"password" json:"-"`
	Var1      string    `db:"var1" json:"-"`
	Var2      string    `db:"var2" json:"-"`
	ApiKey    string    `db:"api_key" json:"-"`
	ApiSecret string    `db:"api_secret" json:"-"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	Limits    int64     `db:"limits" json:"-"`
	Removable bool      `db:"removable" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Form      []NotificationForm
	Routine   chan struct{}
}

type Notifier interface {
	Init() error
	Install() error
	Run() error
	OnFailure(map[string]interface{}) error
	OnSuccess(map[string]interface{}) error
	Select() *Notification
	Test() error
}

type NotificationForm struct {
	id          int64
	Type        string
	Title       string
	Placeholder string
	DbField     string
}

func (n *Notification) isInDatabase() (bool, error) {
	return Collections.Find("id", n.Id).Exists()
}

func SelectNotification(id int64) (*Notification, error) {
	var notifier *Notification
	err := Collections.Find("id", id).One(&notifier)
	return notifier, err
}

func (n *Notification) Update() (*Notification, error) {
	n.CreatedAt = time.Now()
	err := Collections.Find("id", n.Id).Update(n)

	return n, err
}

func InsertDatabase(n *Notification) (int64, error) {
	n.CreatedAt = time.Now()
	newId, err := Collections.Insert(n)
	if err != nil {
		return 0, err
	}
	return newId.(int64), err
}

func Select(id int64) *Notification {
	var notifier *Notification
	for _, n := range AllCommunications {
		notif := n.(Notifier)
		notifier = notif.Select()
		if notifier.Id == id {
			return notifier
		}
	}
	return notifier
}

func SelectNotifier(id int64) Notifier {
	var notifier Notifier
	for _, n := range AllCommunications {
		notif := n.(Notifier)
		n := notif.Select()
		if n.Id == id {
			return notif
		}
	}
	return notifier
}

func (f NotificationForm) Value() string {
	notifier := Select(f.id)
	return notifier.GetValue(f.DbField)
}

func (n *Notification) GetValue(dbField string) string {
	dbField = strings.ToLower(dbField)
	switch dbField {
	case "host":
		return n.Host
	case "port":
		return fmt.Sprintf("%v", n.Port)
	case "username":
		return n.Username
	case "password":
		if n.Password != "" {
			return "##########"
		}
	case "var1":
		return n.Var1
	case "var2":
		return n.Var2
	case "api_key":
		return n.ApiKey
	case "api_secret":
		return n.ApiSecret
	}
	return ""
}

func OnFailure(data map[string]interface{}) {
	for _, comm := range AllCommunications {
		n := comm.(Notifier)
		n.OnFailure(data)
	}
}

func OnSuccess(data map[string]interface{}) {
	for _, comm := range AllCommunications {
		n := comm.(Notifier)
		n.OnSuccess(data)
	}
}

func uniqueMessages(arr []string, v string) []string {
	var newArray []string
	for _, i := range arr {
		if i != v {
			newArray = append(newArray, v)
		}
	}
	return newArray
}
