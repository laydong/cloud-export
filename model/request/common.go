package request

type SourceHTTP struct {
	URL    string                 `json:"url" form:"url" uri:"url" binding:"required"`
	Method string                 `json:"method" form:"method" uri:"method" binding:"required"`
	Header map[string]string      `json:"header" form:"header" uri:"header"`
	Param  map[string]interface{} `json:"param" form:"param" uri:"param"`
}

type ExpSHttpParam struct {
	Timestamp  int64      `json:"timestamp" form:"timestamp" uri:"timestamp" binding:"required"`
	EXTType    string     `json:"ext_type" form:"ext_type" uri:"ext_type" binding:"required"`
	Title      string     `json:"title" form:"title" uri:"title" binding:"required"`
	CallBack   string     `json:"callback" form:"callback" uri:"callback"`
	SourceHTTP SourceHTTP `json:"source_http" form:"source_http" uri:"source_http" binding:"required"`
}
