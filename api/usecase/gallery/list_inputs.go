package gallery

type FolderListInput struct {
	Query    string
	IsPublic *bool
	Page     int
	Limit    int
}

type MediaListInput struct {
	FolderID  string
	IsPublic  *bool
	MediaType string
	Query     string
	Page      int
	Limit     int
}
