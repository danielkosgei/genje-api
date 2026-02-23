package models

import (
	"time"

	"github.com/google/uuid"
)

type VotingRecord struct {
	ID           uuid.UUID `json:"id"`
	PoliticianID uuid.UUID `json:"politician_id"`
	BillName     string    `json:"bill_name"`
	BillNumber   *string   `json:"bill_number,omitempty"`
	Vote         string    `json:"vote"`
	VoteDate     time.Time `json:"vote_date"`
	Session      *string   `json:"session,omitempty"`
	SourceURL    *string   `json:"source_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type ParliamentaryAttendance struct {
	ID           uuid.UUID `json:"id"`
	PoliticianID uuid.UUID `json:"politician_id"`
	SessionDate  time.Time `json:"session_date"`
	Present      bool      `json:"present"`
	SourceURL    *string   `json:"source_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type AttendanceStats struct {
	TotalSessions int     `json:"total_sessions"`
	Present       int     `json:"present"`
	Absent        int     `json:"absent"`
	AttendanceRate float64 `json:"attendance_rate"`
}
