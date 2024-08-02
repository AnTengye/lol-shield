package tree

type Context struct {
	Data interface{} `json:"data"`
	// request info
	Path   string
	Method string
	Params map[string]string
}
