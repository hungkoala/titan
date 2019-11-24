package titan

/**
 * A class that can be used to represent JSON errors that complies to Vnd.Error without the content type requirements.
 * see RFC 2119.
 */
// make it compatible old micronaut code, it's a subject to change latter
type JsonError struct {
	Message  string                 `json:"message"`
	LogRef   string                 `json:"logref"`
	Path     string                 `json:"path"`
	Links    map[string][]string    `json:"_links"`
	Embedded map[string][]JsonError `json:"_embedded"`
}

// see HttpClientResponseExceptionHandler.java
type DefaultJsonError struct {
	Message  string                 `json:"message"`
	LogRef   string                 `json:"logref"`
	Path     string                 `json:"path"`
	Links    map[string][]string    `json:"_links"`
	Embedded map[string][]JsonError `json:"_embedded"`

	TraceId          string            `json:"traceId"`
	ValidationErrors map[string]string `json:"validationErrors"`
	ServerError      string            `json:"serverError"`
}

type HttpClientResponseError struct {
	Message  string
	Response *Response
	Cause    error
}

func (h *HttpClientResponseError) GetMessage() string {
	return h.Message
}

func (h *HttpClientResponseError) Error() string {
	return h.Message
}
