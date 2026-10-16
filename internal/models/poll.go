package models

// Poll is one published opinion poll with its results. SourceURL is the pollster's own publication.
type Poll struct {
	ID             string       `json:"id"`
	Pollster       string       `json:"pollster"`
	Race           string       `json:"race"`
	FieldworkStart *string      `json:"fieldwork_start"`
	FieldworkEnd   *string      `json:"fieldwork_end"`
	SampleSize     *int         `json:"sample_size"`
	MarginOfError  *float64     `json:"margin_of_error"`
	PublishedOn    string       `json:"published_on"`
	SourceURL      string       `json:"source_url"`
	Results        []PollResult `json:"results"`
}

// PollResult is one candidate's share of the vote in a poll, as a percentage.
type PollResult struct {
	CandidateName string  `json:"candidate_name"`
	Party         *string `json:"party"`
	Share         float64 `json:"share"`
}

// PollRaces lists the races a poll can be about. Anything else is rejected by the database.
var PollRaces = []string{"president", "governor", "senate", "party", "other"}
