package models

// Claim is something a politician said or did that has been checked against a source.
// Rating is one of "confirmed", "unproven" or "contested". SourceURL is required.
type Claim struct {
	ID          string  `json:"id"`
	Claim       string  `json:"claim"`
	Rating      string  `json:"rating"`
	RatedOn     *string `json:"rated_on"`
	SourceTitle *string `json:"source_title"`
	SourceURL   string  `json:"source_url"`
}
