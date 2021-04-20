// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type propertyName string

const (
	// Conditions
	propFrom        propertyName = "from"
	propTo          propertyName = "to"
	propSubject     propertyName = "subject"
	propHasWord     propertyName = "hasTheWord"
	propNotHaveWord propertyName = "doesNotHaveTheWord"
	// Actions
	propLabel                propertyName = "label"
	propMarkAsRead           propertyName = "shouldMarkAsRead"
	propArchive              propertyName = "shouldArchive"
	propNeverSpam            propertyName = "shouldNeverSpam"
	propTrash                propertyName = "shouldTrash"
	propNeverMarkAsImportant propertyName = "shouldNeverMarkAsImportant"
	// Ignored
	propOperator propertyName = "sizeOperator"
	propUnit     propertyName = "sizeUnit"
)

var csvHeaders = []string{
	"From",
	"To",
	"Subject",
	"HasWord",
	"NotHaveWord",
	"Labels",
	"MarkAsRead",
	"Archive",
	"NeverSpam",
	"Trash",
	"NeverMarkAsImportant",
}

type root struct {
	Title string `xml:"title"`
	//ID      string `xml:"id"`
	//Updated string `xml:"updated"`
	//Author Author `xml:"author"`
	Entries []entry `xml:"entry"`
}

func (r *root) convertFrom(g *gmailFilters) error {
	r.Entries = make([]entry, len(g.Filters))
	for i := range g.Filters {
		if err := r.Entries[i].convertFrom(&g.Filters[i]); err != nil {
			return err
		}
	}
	return nil
}

type entry struct {
	// Category
	// Title
	ID string `xml:"id"`
	// Updated
	// Content
	Properties []property `xml:"http://schemas.google.com/apps/2006 property"`
}

func (e *entry) addProp(name propertyName, value string) {
	if value == "" {
		return
	}
	e.Properties = append(e.Properties, property{Name: name, Value: value})
}

func (e *entry) setProp(name propertyName, value bool) {
	if !value {
		return
	}
	e.Properties = append(e.Properties, property{Name: name, Value: "true"})
}

func (e *entry) convertFrom(f *filter) error {
	// Conditions
	e.addProp(propFrom, f.From)
	e.addProp(propTo, f.To)
	e.addProp(propSubject, f.Subject)
	e.addProp(propHasWord, f.HasWord.String())
	e.addProp(propNotHaveWord, f.NotHaveWord)
	// Actions
	for _, l := range f.Labels {
		e.addProp(propLabel, l)
	}
	e.setProp(propMarkAsRead, f.MarkAsRead)
	e.setProp(propArchive, f.Archive)
	e.setProp(propNeverSpam, f.NeverSpam)
	e.setProp(propTrash, f.Trash)
	e.setProp(propNeverMarkAsImportant, f.NeverMarkAsImportant)
	return nil
}

type property struct {
	Name  propertyName `xml:"name,attr"`
	Value string       `xml:"value,attr"`
}

type filter struct {
	// Conditions
	From        string          `json:",omitempty"`
	To          string          `json:",omitempty"`
	Subject     string          `json:",omitempty"`
	HasWord     logicExpression `json:",omitempty"`
	NotHaveWord string          `json:",omitempty"`

	// Actions
	Labels               []string `json:",omitempty"`
	MarkAsRead           bool     `json:",omitempty"`
	Archive              bool     `json:",omitempty"`
	NeverSpam            bool     `json:",omitempty"`
	Trash                bool     `json:",omitempty"`
	NeverMarkAsImportant bool     `json:",omitempty"`

	//SizeOperator
	//SizeUnit
}

func (f *filter) toStrings() []string {
	return []string{
		f.From,
		f.To,
		f.Subject,
		f.HasWord.String(),
		f.NotHaveWord,
		strings.Join(f.Labels, ","),
		boolCSV(f.MarkAsRead),
		boolCSV(f.Archive),
		boolCSV(f.NeverSpam),
		boolCSV(f.Trash),
		boolCSV(f.NeverMarkAsImportant),
	}
}

func (f *filter) match(g *filter) bool {
	return f.From == g.From && f.To == g.To && f.Subject == g.Subject && f.HasWord.match(&g.HasWord) && f.NotHaveWord == g.NotHaveWord
}

func (f *filter) less(g *filter) bool {
	// Order by actions.
	if len(f.Labels) != len(g.Labels) {
		return len(f.Labels) < len(g.Labels)
	}
	if len(f.Labels) != 0 {
		for i := range f.Labels {
			if f.Labels[i] != g.Labels[i] {
				return f.Labels[i] < g.Labels[i]
			}
		}
	}
	if f.MarkAsRead != g.MarkAsRead {
		return !f.MarkAsRead
	}
	if f.Archive != g.Archive {
		return !f.Archive
	}
	if f.NeverSpam != g.NeverSpam {
		return !f.NeverSpam
	}
	if f.Trash != g.Trash {
		return !f.Trash
	}
	if f.NeverMarkAsImportant != g.NeverMarkAsImportant {
		return !f.NeverMarkAsImportant
	}
	// Order by conditions.
	if f.From != g.From {
		return f.From < g.From
	}
	if f.To != g.To {
		return f.To < g.To
	}
	if f.Subject != g.Subject {
		return f.Subject < g.Subject
	}
	if f.HasWord.String() != g.HasWord.String() {
		return f.HasWord.String() < g.HasWord.String()
	}
	if f.NotHaveWord != g.NotHaveWord {
		return f.NotHaveWord < g.NotHaveWord
	}
	return false
}

func (f *filter) convertFrom(e *entry) error {
	for _, p := range e.Properties {
		switch p.Name {
		// Conditions
		case propFrom:
			f.From = p.Value
		case propTo:
			f.To = p.Value
		case propSubject:
			f.Subject = p.Value
		case propHasWord:
			if err := f.HasWord.from(p.Value); err != nil {
				return err
			}
		case propNotHaveWord:
			f.NotHaveWord = p.Value

			// Actions
		case propLabel:
			f.Labels = append(f.Labels, p.Value)
			sort.Strings(f.Labels)
		case propMarkAsRead:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.MarkAsRead = true
		case propArchive:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Archive = true
		case propNeverSpam:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.NeverSpam = true
		case propTrash:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Trash = true
		case propNeverMarkAsImportant:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.NeverMarkAsImportant = true

			// Ignored
		case propOperator:
		case propUnit:

		default:
			return fmt.Errorf("unknown property %q", p.Name)
		}
	}
	return nil
}

type logicExpression struct {
	Or []string
}

func (l *logicExpression) match(m *logicExpression) bool {
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
	return nil
}

func (l *logicExpression) String() string {
	return strings.Join(l.Or, " OR ")
}

type gmailFilters struct {
	Filters []filter
}

func (g *gmailFilters) convertFrom(f *root) error {
	g.Filters = make([]filter, len(f.Entries))
	for i := range f.Entries {
		if err := g.Filters[i].convertFrom(&f.Entries[i]); err != nil {
			return err
		}
	}
	return nil
}

// expand expands the OR clauses into single filters.
func (g *gmailFilters) expand(h *gmailFilters) {
	for _, f := range h.Filters {
		if f.To != "" || f.From != "" || f.Subject != "" {
			g.Filters = append(g.Filters, f)
			continue
		}
		f2 := f
		f2.HasWord.Or = nil
		for _, o := range f.HasWord.Or {
			f2.HasWord.Or = []string{o}
			g.Filters = append(g.Filters, f2)
		}
	}
	sort.Slice(g.Filters, func(i, j int) bool {
		return g.Filters[i].less(&g.Filters[j])
	})
}

// compact reduces the number of clauses as much as possible.
func (g *gmailFilters) compact(h *gmailFilters) {
	// Look for redundancy.
	/*
		for _, f := range h.Filters {
			if f.match(nil) {
			}
		}
	*/
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
