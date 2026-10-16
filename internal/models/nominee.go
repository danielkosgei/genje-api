package models

// Nominee is one name from a published party nominee list (IEBC).
// Heading is the list heading as printed in the source document.
type Nominee struct {
	ID         string  `json:"id"`
	Source     string  `json:"source"`
	Party      string  `json:"party"`
	Heading    string  `json:"heading"`
	Name       string  `json:"name"`
	County     *string `json:"county"`
	SourcePage int     `json:"source_page"`
}

// NomineeFilter narrows the nominee list. Empty fields do not filter.
type NomineeFilter struct {
	Party   string
	County  string
	Heading string
	Query   string
	Limit   int
	Offset  int
}

// NomineeHeading is one distinct heading with how many nominees sit under it.
type NomineeHeading struct {
	Heading string `json:"heading"`
	Count   int    `json:"count"`
}
