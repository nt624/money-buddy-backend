package repositories

func defaultStatus(s string) string {
	if s == "planned" || s == "confirmed" {
		return s
	}
	return "confirmed"
}
