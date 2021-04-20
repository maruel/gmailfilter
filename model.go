// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"io"
	"sort"
	"strings"
)

type filter struct {
	Match   match
	Actions actions
}

func (f *filter) toStrings() []string {
	return append(append([]string(nil), f.Match.toStrings()...), f.Actions.toStrings()...)
}

func (f *filter) less(g *filter) bool {
	if f.Actions.less(&g.Actions) {
		return true
	}
	if g.Actions.less(&f.Actions) {
		return false
	}
	return f.Match.less(&g.Match)
}

type match struct {
	From        string          `json:",omitempty"`
	To          string          `json:",omitempty"`
	Subject     string          `json:",omitempty"`
	HasWord     logicExpression `json:",omitempty"`
	NotHaveWord string          `json:",omitempty"` // TODO(maruel): logicExpression
}

func (m *match) String() string {
	out := ""
	if m.From != "" {
		out = "from:(" + m.From + ")"
	}
	if m.To != "" {
		out += " to:(" + m.To + ")"
	}
	if m.Subject != "" {
		out += " to:(" + m.Subject + ")"
	}
	if s := m.HasWord.String(); s != "" {
		out += " " + s
	}
	if m.NotHaveWord != "" {
		out += " -(" + m.NotHaveWord + ")"
	}
	return strings.TrimSpace(out)
}

func (m *match) toStrings() []string {
	return []string{
		m.From,
		m.To,
		m.Subject,
		m.HasWord.String(),
		m.NotHaveWord,
	}
}

func (m *match) equal(n *match) bool {
	return m.From == n.From && m.To == n.To && m.Subject == n.Subject && m.HasWord.equal(&n.HasWord) && m.NotHaveWord == n.NotHaveWord
}

func (m *match) less(n *match) bool {
	if m.From != n.From {
		return m.From < n.From
	}
	if m.To != n.To {
		return m.To < n.To
	}
	if m.Subject != n.Subject {
		return m.Subject < n.Subject
	}
	if m.HasWord.String() != n.HasWord.String() {
		return m.HasWord.String() < n.HasWord.String()
	}
	if m.NotHaveWord != n.NotHaveWord {
		return m.NotHaveWord < n.NotHaveWord
	}
	return false
}

type actions struct {
	Labels         []string `json:",omitempty"`
	MarkAsRead     bool     `json:",omitempty"`
	Archive        bool     `json:",omitempty"`
	NeverSpam      bool     `json:",omitempty"`
	Trash          bool     `json:",omitempty"`
	NeverImportant bool     `json:",omitempty"`

	//SizeOperator
	//SizeUnit
}

func (a *actions) toStrings() []string {
	return []string{
		strings.Join(a.Labels, ","),
		boolCSV(a.MarkAsRead),
		boolCSV(a.Archive),
		boolCSV(a.NeverSpam),
		boolCSV(a.Trash),
		boolCSV(a.NeverImportant),
	}
}

func (a *actions) equal(b *actions) bool {
	if len(a.Labels) != len(b.Labels) {
		return false
	}
	for i := range a.Labels {
		if a.Labels[i] != b.Labels[i] {
			return false
		}
	}
	return a.MarkAsRead == b.MarkAsRead && a.Archive == b.Archive && a.NeverSpam == b.NeverSpam && a.Trash == b.Trash && a.NeverImportant == b.NeverImportant
}

func (a *actions) less(b *actions) bool {
	// Order by actions.
	if len(a.Labels) != len(b.Labels) {
		return len(a.Labels) < len(b.Labels)
	}
	if len(a.Labels) != 0 {
		for i := range a.Labels {
			if a.Labels[i] != b.Labels[i] {
				return a.Labels[i] < b.Labels[i]
			}
		}
	}
	if a.MarkAsRead != b.MarkAsRead {
		return !a.MarkAsRead
	}
	if a.Archive != b.Archive {
		return !a.Archive
	}
	if a.NeverSpam != b.NeverSpam {
		return !a.NeverSpam
	}
	if a.Trash != b.Trash {
		return !a.Trash
	}
	if a.NeverImportant != b.NeverImportant {
		return !a.NeverImportant
	}
	return false
}

type logicExpression struct {
	Or []string
}

func (l *logicExpression) equal(m *logicExpression) bool {
	if len(l.Or) != len(m.Or) {
		return false
	}
	for i := range l.Or {
		if l.Or[i] != m.Or[i] {
			return false
		}
	}
	return true
}

func (l *logicExpression) from(s string) error {
	// Warning: it's very basic and will make errors on complex expressions!
	// - Process () except for list:(...)
	// - Split on " OR "
	l.Or = strings.Split(s, " OR ")
	// An exact list match looks like: list:(<foo.domain.com>)
	return nil
}

func (l *logicExpression) String() string {
	return strings.Join(l.Or, " OR ")
}

type gmailFilters struct {
	Filters []filter
}

// expand expands the OR clauses into single filters.
func (g *gmailFilters) expand(h *gmailFilters) {
	for _, f := range h.Filters {
		if f.Match.To != "" || f.Match.From != "" || f.Match.Subject != "" || f.Match.NotHaveWord != "" {
			g.Filters = append(g.Filters, f)
			continue
		}
		f2 := f
		f2.Match.HasWord.Or = nil
		for _, o := range f.Match.HasWord.Or {
			f2.Match.HasWord.Or = []string{o}
			g.Filters = append(g.Filters, f2)
		}
	}
	sort.Slice(g.Filters, func(i, j int) bool {
		return g.Filters[i].less(&g.Filters[j])
	})
}

// compact reduces the number of clauses as much as possible.
func (g *gmailFilters) compact(h *gmailFilters) {
	// Look for redundancy. Assumes items are sorted, otherwise it'd be O(n^2).
	for _, f := range h.Filters {
		if f.Match.equal(nil) || f.Actions.equal(nil) {
			return
		}
	}
}

func (g *gmailFilters) toCSV(w io.Writer) error {
	c := csv.NewWriter(w)
	if err := c.Write(csvHeaders); err != nil {
		return err
	}
	for _, i := range g.Filters {
		if err := c.Write(i.toStrings()); err != nil {
			return err
		}
	}
	c.Flush()
	return c.Error()
}

func boolCSV(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}
