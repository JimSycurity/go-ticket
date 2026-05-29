package ticket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	MaxTicketFileBytes = 1 << 20
	MaxNoteBytes       = 64 << 10
	MaxSettingsBytes   = 4 << 10
)

const SettingsFileName = "settings.json"

type Ticket struct {
	ID          string
	Status      string
	Deps        []string
	Links       []string
	Created     string
	Type        string
	Priority    string
	Assignee    string
	ExternalRef string
	Parent      string
	Tags        []string
	Unknown     []Field
	Title       string
	Body        string
	Path        string
	RelPath     string
}

type Field struct {
	Key   string
	Value string
}

type Warning struct {
	Path string
	Err  error
}

type Settings struct {
	Prefix string `json:"prefix"`
}

func (w Warning) Error() string {
	return fmt.Sprintf("%s: %v", w.Path, w.Err)
}

var (
	ErrAmbiguousID = errors.New("ambiguous ticket ID")
	ErrMissingID   = errors.New("ticket not found")
)

func ParseFile(root Root, path string) (Ticket, error) {
	data, err := ReadRawFile(path)
	if err != nil {
		return Ticket{}, err
	}
	ticket, err := Parse(root, path, string(data))
	if err != nil {
		return Ticket{}, err
	}
	return ticket, nil
}

func Parse(root Root, path string, content string) (Ticket, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return Ticket{}, fmt.Errorf("missing YAML frontmatter")
	}
	rest := strings.TrimPrefix(normalized, "---\n")
	parts := strings.SplitN(rest, "\n---\n", 2)
	if len(parts) != 2 {
		return Ticket{}, fmt.Errorf("unterminated YAML frontmatter")
	}

	var t Ticket
	t.Path = path
	if root.ProjectDir != "" {
		if rel, err := filepath.Rel(root.ProjectDir, path); err == nil {
			t.RelPath = rel
		}
	}
	if t.RelPath == "" {
		t.RelPath = filepath.Base(path)
	}

	var err error
	for _, line := range strings.Split(parts[0], "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Ticket{}, fmt.Errorf("invalid frontmatter line %q", line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "id":
			t.ID = value
		case "status":
			t.Status = value
		case "deps":
			t.Deps, err = parseList(value)
			if err != nil {
				return Ticket{}, fmt.Errorf("parse deps: %w", err)
			}
		case "links":
			t.Links, err = parseList(value)
			if err != nil {
				return Ticket{}, fmt.Errorf("parse links: %w", err)
			}
		case "created":
			t.Created = value
		case "type":
			t.Type = value
		case "priority":
			t.Priority = value
		case "assignee":
			t.Assignee = value
		case "external-ref":
			t.ExternalRef = value
		case "parent":
			t.Parent = value
		case "tags":
			t.Tags, err = parseList(value)
			if err != nil {
				return Ticket{}, fmt.Errorf("parse tags: %w", err)
			}
		default:
			t.Unknown = append(t.Unknown, Field{Key: key, Value: value})
		}
	}
	if t.ID == "" {
		return Ticket{}, fmt.Errorf("missing required id")
	}
	if _, err := ResolveTicketPath(root, t.ID, false); err != nil {
		return Ticket{}, err
	}

	t.Body = strings.TrimPrefix(parts[1], "\n")
	t.Title = titleFromBody(t.Body)
	return t, nil
}

func List(root Root) ([]Ticket, []Warning) {
	entries, err := os.ReadDir(root.TicketsDir)
	if err != nil {
		return nil, []Warning{{Path: root.TicketsDir, Err: err}}
	}
	var tickets []Ticket
	var warnings []Warning
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		path := filepath.Join(root.TicketsDir, name)
		info, err := os.Lstat(path)
		if err != nil {
			warnings = append(warnings, Warning{Path: path, Err: err})
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			warnings = append(warnings, Warning{Path: path, Err: fmt.Errorf("not a regular ticket file")})
			continue
		}
		ticket, err := ParseFile(root, path)
		if err != nil {
			warnings = append(warnings, Warning{Path: path, Err: err})
			continue
		}
		tickets = append(tickets, ticket)
	}
	sort.Slice(tickets, func(i, j int) bool {
		if tickets[i].Priority != tickets[j].Priority {
			return tickets[i].Priority < tickets[j].Priority
		}
		return tickets[i].ID < tickets[j].ID
	})
	return tickets, warnings
}

func Resolve(root Root, ref string) (Ticket, error) {
	if ref == "" {
		return Ticket{}, ErrMissingID
	}
	entries, err := os.ReadDir(root.TicketsDir)
	if err != nil {
		return Ticket{}, err
	}
	refKey := strings.ToLower(ref)
	var matches []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), ".md")
		if strings.HasPrefix(strings.ToLower(base), refKey) {
			matches = append(matches, filepath.Join(root.TicketsDir, entry.Name()))
		}
	}
	if len(matches) == 0 {
		return Ticket{}, fmt.Errorf("%w: %s", ErrMissingID, ref)
	}
	if len(matches) > 1 {
		return Ticket{}, fmt.Errorf("%w: %s matches %s", ErrAmbiguousID, ref, strings.Join(matchIDs(matches), ", "))
	}
	if _, err := checkedRegularFile(matches[0]); err != nil {
		return Ticket{}, err
	}
	return ParseFile(root, matches[0])
}

func Write(root Root, t Ticket) error {
	path, err := ResolveTicketPath(root, t.ID, false)
	if err != nil {
		return err
	}
	content, err := RenderForWrite(t)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(root.TicketsDir, "."+t.ID+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := replaceFile(tmpPath, path); err != nil {
		return err
	}
	return nil
}

func PreflightWrite(root Root, t Ticket) error {
	if _, err := ResolveTicketPath(root, t.ID, false); err != nil {
		return err
	}
	_, err := RenderForWrite(t)
	return err
}

func RenderForWrite(t Ticket) (string, error) {
	if err := ValidateForWrite(t); err != nil {
		return "", err
	}
	content := Render(t)
	if len([]byte(content)) > MaxTicketFileBytes {
		return "", fmt.Errorf("rendered ticket exceeds %d byte limit: %s", MaxTicketFileBytes, t.ID)
	}
	return content, nil
}

func ReadRawFile(path string) ([]byte, error) {
	info, err := checkedRegularFile(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > MaxTicketFileBytes {
		return nil, fmt.Errorf("ticket file exceeds %d byte limit: %s", MaxTicketFileBytes, path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, MaxTicketFileBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxTicketFileBytes {
		return nil, fmt.Errorf("ticket file exceeds %d byte limit: %s", MaxTicketFileBytes, path)
	}
	return data, nil
}

func ValidateForWrite(t Ticket) error {
	if err := ValidateID(t.ID); err != nil {
		return err
	}
	for key, value := range map[string]string{
		"status":       t.Status,
		"created":      t.Created,
		"type":         t.Type,
		"priority":     t.Priority,
		"assignee":     t.Assignee,
		"external-ref": t.ExternalRef,
		"parent":       t.Parent,
	} {
		if err := validateFrontmatterScalar(key, value); err != nil {
			return err
		}
	}
	if t.Parent != "" {
		if err := ValidateID(t.Parent); err != nil {
			return fmt.Errorf("invalid parent: %w", err)
		}
	}
	for _, list := range []struct {
		name   string
		values []string
	}{
		{name: "deps", values: t.Deps},
		{name: "links", values: t.Links},
		{name: "tags", values: t.Tags},
	} {
		for _, value := range list.values {
			if err := validateListAtom(list.name, value); err != nil {
				return err
			}
		}
	}
	for _, field := range t.Unknown {
		if err := validateFieldKey(field.Key); err != nil {
			return err
		}
		if err := validateFrontmatterScalar(field.Key, field.Value); err != nil {
			return err
		}
	}
	return nil
}

func Render(t Ticket) string {
	var b strings.Builder
	b.WriteString("---\n")
	writeField(&b, "id", t.ID)
	writeField(&b, "status", valueOr(t.Status, "open"))
	writeList(&b, "deps", t.Deps)
	writeList(&b, "links", t.Links)
	writeField(&b, "created", valueOr(t.Created, time.Now().UTC().Format(time.RFC3339)))
	writeField(&b, "type", valueOr(t.Type, "task"))
	writeField(&b, "priority", valueOr(t.Priority, "2"))
	if t.Assignee != "" {
		writeField(&b, "assignee", t.Assignee)
	}
	if t.ExternalRef != "" {
		writeField(&b, "external-ref", t.ExternalRef)
	}
	if t.Parent != "" {
		writeField(&b, "parent", t.Parent)
	}
	if len(t.Tags) > 0 {
		writeList(&b, "tags", t.Tags)
	}
	for _, field := range t.Unknown {
		if field.Key != "" {
			writeField(&b, field.Key, field.Value)
		}
	}
	b.WriteString("---\n")
	body := strings.ReplaceAll(t.Body, "\r\n", "\n")
	if body == "" {
		body = "# " + t.Title + "\n"
	}
	if !strings.HasPrefix(body, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func NewTicket(root Root, title string) (Ticket, error) {
	id, err := GenerateID(root)
	if err != nil {
		return Ticket{}, err
	}
	return Ticket{
		ID:       id,
		Status:   "open",
		Deps:     []string{},
		Links:    []string{},
		Created:  time.Now().UTC().Format(time.RFC3339),
		Type:     "task",
		Priority: "2",
		Title:    title,
		Body:     "# " + title + "\n",
	}, nil
}

func GenerateID(root Root) (string, error) {
	prefix := projectPrefix(root.ProjectDir)
	settings, err := LoadSettings(root)
	if err != nil {
		return "", err
	}
	if settings.Prefix != "" {
		prefix = settings.Prefix
	}
	for i := 0; i < 100; i++ {
		suffix, err := randomSuffix()
		if err != nil {
			return "", err
		}
		id := prefix + "-" + suffix
		if _, err := ResolveTicketPath(root, id, true); os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return id, nil
		}
		if _, err := Resolve(root, id); errors.Is(err, ErrMissingID) {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not generate non-colliding ticket ID")
}

func LoadSettings(root Root) (Settings, error) {
	path := filepath.Join(root.TicketsDir, SettingsFileName)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Settings{}, nil
		}
		return Settings{}, fmt.Errorf("stat ticket settings: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Settings{}, fmt.Errorf("ticket settings is not a regular file: %s", path)
	}
	if info.Size() > MaxSettingsBytes {
		return Settings{}, fmt.Errorf("ticket settings exceeds %d byte limit: %s", MaxSettingsBytes, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, fmt.Errorf("read ticket settings: %w", err)
	}
	var settings Settings
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&settings); err != nil {
		return Settings{}, fmt.Errorf("parse ticket settings: %w", err)
	}
	if settings.Prefix != "" {
		if err := validatePrefix(settings.Prefix); err != nil {
			return Settings{}, err
		}
	}
	return settings, nil
}

func validatePrefix(prefix string) error {
	if err := ValidateID(prefix); err != nil {
		return fmt.Errorf("invalid settings prefix: %w", err)
	}
	if strings.Contains(prefix, "-") {
		return fmt.Errorf("invalid settings prefix %q: use an ID atom without hyphen", prefix)
	}
	return nil
}

func AddUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func RemoveValue(values []string, value string) []string {
	var out []string
	for _, existing := range values {
		if existing != value {
			out = append(out, existing)
		}
	}
	return out
}

func AppendNote(t Ticket, text string, now time.Time) Ticket {
	text = strings.TrimSpace(text)
	if text == "" {
		return t
	}
	body := strings.TrimRight(strings.ReplaceAll(t.Body, "\r\n", "\n"), "\n")
	if !strings.Contains(body, "\n## Notes\n") && !strings.HasSuffix(body, "\n## Notes") {
		body += "\n\n## Notes"
	}
	body += "\n\n**" + now.UTC().Format(time.RFC3339) + "**\n\n" + text + "\n"
	t.Body = body
	return t
}

func ValidateID(id string) error {
	if !ticketIDPattern.MatchString(id) {
		return fmt.Errorf("invalid ticket ID %q: use letters, numbers, underscore, or hyphen only", id)
	}
	if _, reserved := windowsReservedNames[strings.ToUpper(id)]; reserved {
		return fmt.Errorf("invalid ticket ID %q: reserved Windows device name", id)
	}
	return nil
}

func IsWorkStatus(status string) bool {
	return status == "open" || status == "in_progress"
}

func IsValidStatus(status string) bool {
	return status == "open" || status == "in_progress" || status == "closed"
}

func IsValidType(ticketType string) bool {
	switch ticketType {
	case "bug", "feature", "task", "epic", "chore":
		return true
	default:
		return false
	}
}

func IsValidPriority(priority string) bool {
	return priority == "0" || priority == "1" || priority == "2" || priority == "3" || priority == "4"
}

func parseList(value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return []string{}, nil
	}
	if strings.HasPrefix(value, "[") != strings.HasSuffix(value, "]") {
		return nil, fmt.Errorf("malformed inline list %q", value)
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		value = strings.TrimSuffix(strings.TrimPrefix(value, "["), "]")
	}
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.Trim(strings.TrimSpace(part), `"'`)
		if part != "" {
			out = append(out, part)
		}
	}
	return out, nil
}

func titleFromBody(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func writeField(b *strings.Builder, key string, value string) {
	b.WriteString(key)
	b.WriteString(": ")
	b.WriteString(value)
	b.WriteByte('\n')
}

func writeList(b *strings.Builder, key string, values []string) {
	b.WriteString(key)
	b.WriteString(": [")
	b.WriteString(strings.Join(values, ", "))
	b.WriteString("]\n")
}

func valueOr(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func projectPrefix(projectDir string) string {
	name := strings.ToLower(filepath.Base(projectDir))
	var parts []string
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || unicode.IsSpace(r)
	}) {
		clean := cleanToken(part)
		if clean != "" {
			parts = append(parts, clean)
		}
	}
	if len(parts) > 1 {
		var b strings.Builder
		for _, part := range parts {
			b.WriteByte(part[0])
		}
		return b.String()
	}
	clean := cleanToken(name)
	if len(clean) >= 3 {
		return clean[:3]
	}
	if clean != "" {
		return clean
	}
	return "tk"
}

func cleanToken(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func randomSuffix() (string, error) {
	var buf [3]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:])[:4], nil
}

func matchIDs(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, strings.TrimSuffix(filepath.Base(path), ".md"))
	}
	sort.Strings(out)
	return out
}

func checkedRegularFile(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("ticket file is a symlink: %s", path)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("ticket path is not a regular file: %s", path)
	}
	return info, nil
}

func validateFieldKey(key string) error {
	if key == "" {
		return fmt.Errorf("frontmatter key cannot be empty")
	}
	for _, r := range key {
		if r == '\r' || r == '\n' || r == ':' || unicode.IsSpace(r) {
			return fmt.Errorf("invalid frontmatter key %q", key)
		}
	}
	return nil
}

func validateFrontmatterScalar(key string, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("frontmatter field %q contains a newline", key)
	}
	return nil
}

func validateListAtom(key string, value string) error {
	if value == "" {
		return fmt.Errorf("frontmatter list %q contains an empty value", key)
	}
	if strings.ContainsAny(value, "\r\n,[]") {
		return fmt.Errorf("frontmatter list %q contains unsafe value %q", key, value)
	}
	return nil
}
