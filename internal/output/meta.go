package output

// Meta is an optional envelope metadata section intended for agents.
//
// Schema: optional $id of the JSON schema that `data` conforms to.
// Status: HTTP status code (when known).
type Meta struct {
	Schema string `json:"schema,omitempty"`
	Status int    `json:"status,omitempty"`
}
