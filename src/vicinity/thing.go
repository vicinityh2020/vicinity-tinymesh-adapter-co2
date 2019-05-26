package vicinity

type Schema struct {
	Type string `json:"type"`
}

type Field struct {
	Name   string `json:"name"`
	Schema Schema `json:"schema"`
}

type IO struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Fields      []Field `json:"field"`
}

type Event struct {
	Eid      string `json:"eid"`
	Monitors string `json:"monitors"`
	Output   IO     `json:"output"`
}

type Link struct {
	Href   string `json:"href"`
	Output IO     `json:"output"`
}

type Property struct {
	Pid      string `json:"pid"`
	Monitors string `json:"monitors"`
	ReadLink Link   `json:"read_link"`
}

type Device struct {
	Oid        string        `json:"oid"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Version    string        `json:"version"`
	Keywords   []string      `json:"keywords"`
	Properties []Property    `json:"properties"`
	Actions    []interface{} `json:"actions"`
	Events     []Event       `json:"events"`
}
