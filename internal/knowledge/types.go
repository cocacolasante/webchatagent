package knowledge

// BusinessInfo holds core business details used in the system prompt.
type BusinessInfo struct {
	Name     string        `json:"name"`
	Tagline  string        `json:"tagline"`
	Industry string        `json:"industry"`
	Website  string        `json:"website"`
	Email    string        `json:"email"`
	Phone    string        `json:"phone"`
	Address  string        `json:"address"`
	Hours    string        `json:"hours"`
	Services []Service     `json:"services"`
	Team     []TeamMember  `json:"team"`
}

// Service represents a business offering.
type Service struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price,omitempty"`
}

// TeamMember represents a staff member.
type TeamMember struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

// FAQ is a question/answer pair.
type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// Document is a rich knowledge document with a title, optional URL, and content.
type Document struct {
	Title   string `json:"title"`
	URL     string `json:"url,omitempty"`
	Content string `json:"content"`
}

// KnowledgeBase is the full knowledge configuration for a tenant.
type KnowledgeBase struct {
	FAQs               []FAQ      `json:"faqs"`
	CustomInstructions string     `json:"customInstructions"`
	Documents          []Document `json:"documents"`
}
