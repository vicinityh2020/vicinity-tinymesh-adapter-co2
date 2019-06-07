package vicinity

type Location struct {
	LocationType string `json:"location_type"`
	LocationId   string `json:"location_id,omitempty"`
	Label        string `json:"label"`
}

type Schema struct {
	Type string `json:"type"`
}

type Field struct {
	Name        string `json:"name"`
	Schema      Schema `json:"schema"`
	Type        string `json:"type,omitempty"`
	Predicate   string `json:"predicate,omitempty"`
	Description string `json:"description,omitempty"`
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
	LocatedIn  []Location    `json:"located-in"`
}
