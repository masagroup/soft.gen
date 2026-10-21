package naming

// AvoidKeywords returns name with suffix appended when name is a key of
// keywords, and name unchanged otherwise: with the Go keywords and the
// suffix "_", "type" gives "type_" and "name" gives "name".
func AvoidKeywords(name string, keywords map[string]bool, suffix string) string {
	if keywords[name] {
		return name + suffix
	}
	return name
}
