package game

import "slices"

// History retains first-publicly-seen identities only. It intentionally contains
// no present controller, zone, restriction marker, binding location or deck index.
func rememberPublic(s *State, id string) {
	if !slices.Contains(s.PublicHistory, id) {
		s.PublicHistory = append(s.PublicHistory, id)
	}
}
func publicZone(z Zone) bool {
	switch z {
	case Series, Attachment, Formation, Unassigned, ActiveAce, Initiative, Discard:
		return true
	}
	return false
}
