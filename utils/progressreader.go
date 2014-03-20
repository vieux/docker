package utils

import (
	"io"
	"time"
)

// Reader with progress bar
type progressReader struct {
	reader     io.ReadCloser // Stream to read from
	output     io.Writer     // Where to send progress bar to
	progress   JSONProgress
	lastUpdate int // How many bytes read at least update
	action     string
	sf         *StreamFormatter
	newLine    bool
	global     *JSONProgress
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	read, err := r.reader.Read(p)
	r.progress.Current += read
	if r.global != nil {
		r.global.Current += read
	}
	updateEvery := 1024 * 512 //512kB
	if r.progress.Total > 0 {
		// Update progress for every 1% read if 1% < 512kB
		if increment := int(0.01 * float64(r.progress.Total)); increment < updateEvery {
			updateEvery = increment
		}
	}
	if r.progress.Current-r.lastUpdate > updateEvery || err != nil {
		r.output.Write(r.sf.FormatProgress(r.progress.key, r.action, &r.progress))
		if r.global != nil {
			r.output.Write(r.sf.FormatProgress(r.global.key, r.action, r.global))
		}
		r.lastUpdate = r.progress.Current
	}
	// Send newline when complete
	if r.newLine && err != nil {
		r.output.Write(r.sf.FormatStatus("", ""))
	}
	return read, err
}
func (r *progressReader) Close() error {
	r.progress.Current = r.progress.Total
	r.output.Write(r.sf.FormatProgress(r.progress.key, r.action, &r.progress))
	return r.reader.Close()
}
func ProgressReader(r io.ReadCloser, size int, output io.Writer, sf *StreamFormatter, newline bool, ID, action string, global *JSONProgress) *progressReader {
	return &progressReader{
		reader:   r,
		output:   NewWriteFlusher(output),
		action:   action,
		progress: JSONProgress{key: ID, Total: size, Start: time.Now().UTC().Unix()},
		sf:       sf,
		newLine:  newline,
		global:   global,
	}
}
