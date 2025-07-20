package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ZoomMeta struct {
	Topic        string
	Start        time.Time
	Host         string
	HostEmail    string
	Participants []ZParticipant
	Files        []ZFile
}

type ZParticipant struct {
	Name  string
	Email string
}

type ZFile struct {
	Name string
	Path string
	Type string // MIME type e.g. "video/mp4"
}

func readZoomMeta(folder string) (*ZoomMeta, error) {
	fi, err := os.Stat(folder)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("zoom: path is not a directory")
	}

	meta := &ZoomMeta{}

	// 1) Try rich side-car -> meeting_info.json  (rare but newest Zoom builds)
	if tryReadMeetingInfoJSON(folder, meta) == nil {
		// success – we already filled the struct
	} else {
		// 2) Try recording.conf (older Zoom builds – only has a/v filenames)
		tryReadRecordingConf(folder, meta)

		// 3) Parse folder name  “YYYY-MM-DD_HH-MM-SS_<topic words>”
		parseFolderName(fi.Name(), meta)
	}

	// 4) Enumerate every artefact (mp4, m4a, vtt, txt …)
	err = filepath.WalkDir(folder, func(path string, d fs.DirEntry, wErr error) error {
		if wErr != nil || d.IsDir() {
			return wErr
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		switch ext {
		case ".mp4", ".m4a", ".mov", ".vtt", ".txt":
			mt := mime.TypeByExtension(ext)
			if mt == "" {
				mt = "application/octet-stream"
			}
			meta.Files = append(meta.Files, ZFile{
				Name: d.Name(),
				Path: path,
				Type: mt,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func tryReadMeetingInfoJSON(folder string, meta *ZoomMeta) error {
	fp := filepath.Join(folder, "meeting_info.json")
	raw, err := os.ReadFile(fp)
	if err != nil {
		return err
	}

	// Zoom appears to use snake-case keys – we unmarshal into a loose map
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}

	if t, ok := m["topic"].(string); ok {
		meta.Topic = t
	}
	if t, ok := m["meeting_title"].(string); ok && meta.Topic == "" {
		meta.Topic = t
	}
	if h, ok := m["host_name"].(string); ok {
		meta.Host = h
	}
	if e, ok := m["host_email"].(string); ok {
		meta.HostEmail = e
	}
	if ts, ok := m["start_time"].(string); ok {
		if tt, _ := time.Parse(time.RFC3339, ts); !tt.IsZero() {
			meta.Start = tt
		}
	}
	// participants (array of objects w/ name + email)
	if p, ok := m["participants"].([]any); ok {
		for _, v := range p {
			if obj, ok := v.(map[string]any); ok {
				meta.Participants = append(meta.Participants, ZParticipant{
					Name:  str(obj["name"]),
					Email: str(obj["email"]),
				})
			}
		}
	}
	return nil
}

// recording.conf – only lists audio/video filenames
func tryReadRecordingConf(folder string, meta *ZoomMeta) {
	fp := filepath.Join(folder, "recording.conf")
	raw, err := os.ReadFile(fp)
	if err != nil {
		return
	}
	var rc struct {
		Items []struct {
			Audio  string `json:"audio"`
			Video  string `json:"video"`
			Prefix string `json:"prefix"`
		} `json:"items"`
	}
	if json.Unmarshal(raw, &rc) != nil || len(rc.Items) == 0 {
		return
	}
	for _, itm := range rc.Items {
		if itm.Video != "" {
			meta.Files = append(meta.Files, ZFile{
				Name: itm.Video,
				Path: filepath.Join(folder, itm.Video),
				Type: "video/mp4",
			})
		}
		if itm.Audio != "" {
			meta.Files = append(meta.Files, ZFile{
				Name: itm.Audio,
				Path: filepath.Join(folder, itm.Audio),
				Type: "audio/m4a",
			})
		}
	}
}

var folderRe = regexp.MustCompile(`^(?P<date>\d{4}-\d{2}-\d{2})[_ ](?P<time>\d{2}[-.]\d{2}[-.]\d{2})[_ ]?(?P<topic>.*)$`)

func parseFolderName(name string, meta *ZoomMeta) {
	m := folderRe.FindStringSubmatch(name)
	if len(m) == 0 {
		return
	}
	parts := make(map[string]string)
	for i, n := range folderRe.SubexpNames() {
		if i > 0 && n != "" {
			parts[n] = m[i]
		}
	}
	if parts["date"] != "" && parts["time"] != "" {
		layout := "2006-01-02 15-04-05" // default pattern
		ts := parts["date"] + " " + strings.ReplaceAll(parts["time"], ".", "-")
		if t, err := time.Parse(layout, ts); err == nil {
			meta.Start = t
		}
	}
	if meta.Topic == "" && parts["topic"] != "" {
		meta.Topic = strings.ReplaceAll(parts["topic"], "_", " ")
	}
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
