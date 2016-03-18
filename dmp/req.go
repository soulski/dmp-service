package dmp

const (
	REQ_RES = "req-res"
	PUB_SUB = "pub-sub"
)

type Service struct {
	Namespace    string `json:"namespace"`
	ContactPoint string `json:"contact-point"`
}

type Message struct {
	Type      string `json:"type"`
	Topic     string `json:"topic"`
	Namespace string `json:"namespace"`
	Body      string `json:"body"`
}
