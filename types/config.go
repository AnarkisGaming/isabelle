package types

// Configuration specifies the data structures used in config objects
type Configuration struct {
	Discord struct {
		Token  string `json:"token"`
		Output string `json:"output"`
	} `json:"discord"`
	Email struct {
		Host  string `json:"host"`
		User  string `json:"user"`
		Pass  string `json:"pass"`
		Every int    `json:"every"`
		// Caveat emptor! If Isabelle doesn't delete the messages, she will keep tripping over them. Should be used for debugging only.
		Keep bool `json:"keep"`
	} `json:"email"`
	GitHub struct {
		PAT    string   `json:"pat"`
		Repo   string   `json:"repo"`
		Labels []string `json:"labels"`
		React  string   `json:"react"`
	} `json:"github"`
	Messages struct {
		InvalidFile string
		BadZIP      string
		BadDownload string
		Help        string
		Success     string
	} `json:"messages"`
	Files struct {
		File1 string `json:"file1"`
		File2 string `json:"file2"`
	} `json:"files"`
}
