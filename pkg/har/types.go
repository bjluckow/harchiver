package har

// TODO: Maybe good to generate this file with a script using JSON schema .har specs?

// HttpArchive is the top-level HAR object.
type HttpArchive struct {
	Log Log `json:"log"`
}

// required by HAR 1.2 spec
type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Log contains the recorded network activity.
type Log struct {
	Creator Creator `json:"creator"`
	Version string  `json:"version`
	Entries []Entry `json:"entries"`
}

// Entry represents a single HTTP transaction.
type Entry struct {
	StartedDateTime string   `json:"startedDateTime,omitempty"`
	Time            float64  `json:"time,omitempty"`
	Request         Request  `json:"request"`
	Response        Response `json:"response"`
}

// Request models the outgoing HTTP request.
type Request struct {
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	HTTPVersion string    `json:"httpVersion,omitempty"`
	Headers     []Header  `json:"headers,omitempty"`
	QueryString []Query   `json:"queryString,omitempty"`
	Cookies     []Cookie  `json:"cookies,omitempty"`
	PostData    *PostData `json:"postData,omitempty"`
	HeadersSize int64     `json:"headersSize,omitempty"`
	BodySize    int64     `json:"bodySize,omitempty"`
}

// Response models the incoming HTTP response.
type Response struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText,omitempty"`
	HTTPVersion string   `json:"httpVersion,omitempty"`
	Headers     []Header `json:"headers,omitempty"`
	Cookies     []Cookie `json:"cookies,omitempty"`
	Content     *Content `json:"content,omitempty"`
	RedirectURL string   `json:"redirectURL,omitempty"`
	HeadersSize int64    `json:"headersSize,omitempty"`
	BodySize    int64    `json:"bodySize,omitempty"`
}

// Header represents a single HTTP header.
type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Query represents a single query string parameter.
type Query struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Cookie represents a single cookie.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Expires  string `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
}

// PostData represents request body data.
type PostData struct {
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
}

// Content represents response body metadata.
type Content struct {
	Size     int64  `json:"size,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Encoding string `json:"encoding,omitempty"` // e.g. "base64"
}
