package models

import "time"

// Transaction is a minimal, agent-friendly view of a Sure transaction.
// It is shared across packages (api, insights, commands) to avoid map[string]any plumbing.
//
// NOTE: AmountText is kept as returned by Sure (string) because the upstream API currently
// has known quirks around sign formatting; insights normalize via Classification.
type Transaction struct {
	ID             string
	Name           string
	Classification string // income|expense
	AmountText     string // e.g. "€1.00" or "-€2.00"
	Currency       string
	Date           time.Time
	AccountName    string
	MerchantName   string
	CategoryName   string
	CategoryID     string
}
