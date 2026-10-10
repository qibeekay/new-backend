package models

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Metric struct {
	Value string `json:"value" example:"10x"`
	Label string `json:"label" example:"Throughput increase"`
}

// Project mirrors the frontend's `Project` type (types/portfolio.ts).
type Project struct {
	ID          string   `json:"id" example:"ledger-api-lx7y"`
	Title       string   `json:"title" example:"Ledger API"`
	Tagline     string   `json:"tagline" example:"High-throughput financial ledger"`
	Year        string   `json:"year" example:"2024"`
	Category    string   `json:"category" example:"systems" enums:"creative,functional,systems"`
	Status      string   `json:"status" example:"completed" enums:"completed,ongoing,ideating,paused"`
	Featured    bool     `json:"featured,omitempty" example:"true"`
	Image       string   `json:"image,omitempty" example:"https://example.com/cover.jpg"`
	ImageAlt    string   `json:"imageAlt,omitempty" example:"Project preview screenshot"`
	Description string   `json:"description" example:"Designed and implemented an immutable distributed ledger with sub-millisecond query performance."`
	Role        string   `json:"role" example:"Lead Backend Engineer"`
	Stack       []string `json:"stack" example:"Go,PostgreSQL,Docker"`
	Metrics     []Metric `json:"metrics"`
	LiveURL     string   `json:"liveUrl,omitempty" example:"https://example.com"`
	CodeURL     string   `json:"codeUrl,omitempty" example:"https://github.com/example/ledger"`
}

// ProjectInput represents the payload when creating or updating a project.
type ProjectInput struct {
	Title       string   `json:"title" example:"Ledger API"`
	Tagline     string   `json:"tagline" example:"High-throughput financial ledger"`
	Year        string   `json:"year" example:"2024"`
	Category    string   `json:"category" example:"systems" enums:"creative,functional,systems"`
	Status      string   `json:"status" example:"completed" enums:"completed,ongoing,ideating,paused"`
	Featured    bool     `json:"featured,omitempty" example:"true"`
	Image       string   `json:"image,omitempty" example:"https://example.com/cover.jpg"`
	ImageAlt    string   `json:"imageAlt,omitempty" example:"Project preview screenshot"`
	Description string   `json:"description" example:"Designed and implemented an immutable distributed ledger with sub-millisecond query performance."`
	Role        string   `json:"role" example:"Lead Backend Engineer"`
	Stack       []string `json:"stack" example:"Go,PostgreSQL,Docker"`
	Metrics     []Metric `json:"metrics"`
	LiveURL     string   `json:"liveUrl,omitempty" example:"https://example.com"`
	CodeURL     string   `json:"codeUrl,omitempty" example:"https://github.com/example/ledger"`
}

var (
	categories = map[string]bool{"creative": true, "functional": true, "systems": true}
	statuses   = map[string]bool{"completed": true, "ongoing": true, "ideating": true, "paused": true}
	yearRe     = regexp.MustCompile(`^\d{4}$`)
	slugRe     = regexp.MustCompile(`[^a-z0-9]+`)
)

// Normalize trims input and drops empty/duplicate entries.
func (p *Project) Normalize() {
	p.Title = strings.TrimSpace(p.Title)
	p.Tagline = strings.TrimSpace(p.Tagline)
	p.Year = strings.TrimSpace(p.Year)
	p.Description = strings.TrimSpace(p.Description)
	p.Role = strings.TrimSpace(p.Role)
	p.Image = strings.TrimSpace(p.Image)
	p.ImageAlt = strings.TrimSpace(p.ImageAlt)
	p.LiveURL = strings.TrimSpace(p.LiveURL)
	p.CodeURL = strings.TrimSpace(p.CodeURL)

	seen := map[string]bool{}
	stack := make([]string, 0, len(p.Stack))
	for _, s := range p.Stack {
		s = strings.TrimSpace(s)
		k := strings.ToLower(s)
		if s == "" || seen[k] {
			continue
		}
		seen[k] = true
		stack = append(stack, s)
	}
	p.Stack = stack

	metrics := make([]Metric, 0, len(p.Metrics))
	for _, m := range p.Metrics {
		m.Value, m.Label = strings.TrimSpace(m.Value), strings.TrimSpace(m.Label)
		if m.Value != "" && m.Label != "" {
			metrics = append(metrics, m)
		}
	}
	p.Metrics = metrics
}

// Validate returns field -> message.
func (p *Project) Validate() map[string]string {
	errs := map[string]string{}
	if p.Title == "" {
		errs["title"] = "Give the project a name."
	}
	if p.Tagline == "" {
		errs["tagline"] = "A one-line summary is required."
	}
	if !yearRe.MatchString(p.Year) {
		errs["year"] = "Use a four-digit year."
	}
	if len(p.Description) < 20 {
		errs["description"] = "Write at least a couple of sentences."
	}
	if !categories[p.Category] {
		errs["category"] = "Must be creative, functional or systems."
	}
	if !statuses[p.Status] {
		errs["status"] = "Must be completed, ongoing, ideating or paused."
	}
	if len(p.Metrics) > 3 {
		errs["metrics"] = "At most three key results."
	}
	for field, v := range map[string]string{"image": p.Image, "liveUrl": p.LiveURL, "codeUrl": p.CodeURL} {
		if v != "" && !isHTTPURL(v) {
			errs[field] = "Must be a full URL, starting with https://"
		}
	}
	return errs
}

func isHTTPURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func NewProjectID(title string) string {
	slug := strings.Trim(slugRe.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if slug == "" {
		slug = "project"
	}
	return slug + "-" + strconv.FormatInt(time.Now().UnixMilli(), 36)
}
