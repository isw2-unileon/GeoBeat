package genre

// Genre represents a music genre with its ID, name, and normalized name.
type Genre struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NormalizedName string `json:"normalized_name"`
}
