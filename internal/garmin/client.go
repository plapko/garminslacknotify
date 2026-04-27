package garmin

// Activity holds the workout data needed for status formatting.
type Activity struct {
	TypeKey  string
	Duration float64 // seconds
	Distance float64 // metres, 0 if not applicable
}
