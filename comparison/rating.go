package comparison

// Rating represents a user's subjective assessment of a model's output
type Rating struct {
	RhythmicInterest int `json:"rhythmic_interest"` // Syncopation, variety, groove (1-5)
	VelocityDynamics int `json:"velocity_dynamics"` // Accent placement, variation (1-5)
	GenreAccuracy    int `json:"genre_accuracy"`    // Matches requested style (1-5)
	Overall          int `json:"overall"`           // Overall quality assessment (1-5)
}

// ValidRatingCriteria lists all valid criteria names for rating
var ValidRatingCriteria = []string{"rhythmic", "dynamics", "genre", "overall", "all"}

// IsValidCriteria checks if a criteria name is valid
func IsValidCriteria(criteria string) bool {
	for _, c := range ValidRatingCriteria {
		if c == criteria {
			return true
		}
	}
	return false
}

// IsValidScore checks if a score is within the valid range (1-5)
func IsValidScore(score int) bool {
	return score >= 1 && score <= 5
}

// NewRating creates a new rating with all values set to 0 (unrated)
func NewRating() *Rating {
	return &Rating{}
}

// SetCriteria sets a specific criteria score
func (r *Rating) SetCriteria(criteria string, score int) bool {
	switch criteria {
	case "rhythmic":
		r.RhythmicInterest = score
	case "dynamics":
		r.VelocityDynamics = score
	case "genre":
		r.GenreAccuracy = score
	case "overall":
		r.Overall = score
	case "all":
		r.RhythmicInterest = score
		r.VelocityDynamics = score
		r.GenreAccuracy = score
		r.Overall = score
	default:
		return false
	}
	return true
}

// GetCriteria gets a specific criteria score
func (r *Rating) GetCriteria(criteria string) (int, bool) {
	switch criteria {
	case "rhythmic":
		return r.RhythmicInterest, true
	case "dynamics":
		return r.VelocityDynamics, true
	case "genre":
		return r.GenreAccuracy, true
	case "overall":
		return r.Overall, true
	default:
		return 0, false
	}
}

// IsComplete checks if all criteria have been rated (non-zero)
func (r *Rating) IsComplete() bool {
	return r.RhythmicInterest > 0 && r.VelocityDynamics > 0 &&
		r.GenreAccuracy > 0 && r.Overall > 0
}
