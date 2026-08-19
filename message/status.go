package message

// Status represents the alert/notification severity status.
type Status string

const (
	// StatusDefault represents standard/neutral info (gray / default).
	StatusDefault Status = "default"
	// StatusInfo represents informational message (blue).
	StatusInfo Status = "info"
	// StatusSuccess represents success message (green).
	StatusSuccess Status = "success"
	// StatusWarning represents warning message (orange / yellow).
	StatusWarning Status = "warning"
	// StatusDanger represents critical / failure message (red).
	StatusDanger Status = "danger"
)

// LarkTemplate returns the corresponding Lark Card template color tag.
func (s Status) LarkTemplate() string {
	switch s {
	case StatusSuccess:
		return "green"
	case StatusWarning:
		return "orange"
	case StatusDanger:
		return "red"
	case StatusInfo:
		return "blue"
	default:
		return "grey"
	}
}

// IconEmoji returns a standard status emoji indicator.
func (s Status) IconEmoji() string {
	switch s {
	case StatusSuccess:
		return "✅"
	case StatusWarning:
		return "⚠️"
	case StatusDanger:
		return "🚨"
	case StatusInfo:
		return "ℹ️"
	default:
		return "📢"
	}
}
