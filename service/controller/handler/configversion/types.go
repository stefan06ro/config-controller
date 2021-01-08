package configversion

type Index struct {
	APIVersion string                  `json:"apiVersion"`
	Entries    map[string][]IndexEntry `json:"entries"`
}

type IndexEntry struct {
	APIVersion    string   `json:"apiVersion"`
	AppVersion    string   `json:"appVersion"`
	ConfigVersion string   `json:"configVersion,omitempty"`
	Created       string   `json:"created"`
	Description   string   `json:"description"`
	Digest        string   `json:"digest"`
	Home          string   `json:"home"`
	Name          string   `json:"name"`
	Urls          []string `json:"urls"`
	Version       string   `json:"version"`
}
