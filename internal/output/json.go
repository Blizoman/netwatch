package output

import (
	"encoding/json"
	"io"
)

// jsonWriter emits one compact, newline-delimited JSON object per round,
// suitable for piping into jq or a log processor.
type jsonWriter struct {
	w io.Writer
}

func (j *jsonWriter) Write(round Round) error {
	enc := json.NewEncoder(j.w)
	return enc.Encode(round)
}
