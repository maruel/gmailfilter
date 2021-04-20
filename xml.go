// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"sort"
)

type propertyName string

const (
	// Match
	propFrom        propertyName = "from"
	propTo          propertyName = "to"
	propSubject     propertyName = "subject"
	propHasWord     propertyName = "hasTheWord"
	propNotHaveWord propertyName = "doesNotHaveTheWord"
	// Actions
	propLabel          propertyName = "label"
	propMarkAsRead     propertyName = "shouldMarkAsRead"
	propArchive        propertyName = "shouldArchive"
	propNeverSpam      propertyName = "shouldNeverSpam"
	propTrash          propertyName = "shouldTrash"
	propNeverImportant propertyName = "shouldNeverMarkAsImportant"
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
	"NeverImportant",
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

func (r *root) convertTo(g *gmailFilters) error {
	g.Filters = make([]filter, len(r.Entries))
	for i := range r.Entries {
		if err := r.Entries[i].convertTo(&g.Filters[i]); err != nil {
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
	e.addProp(propFrom, f.Match.From)
	e.addProp(propTo, f.Match.To)
	e.addProp(propSubject, f.Match.Subject)
	e.addProp(propHasWord, f.Match.HasWord.String())
	e.addProp(propNotHaveWord, f.Match.NotHaveWord)

	for _, l := range f.Actions.Labels {
		e.addProp(propLabel, l)
	}
	e.setProp(propMarkAsRead, f.Actions.MarkAsRead)
	e.setProp(propArchive, f.Actions.Archive)
	e.setProp(propNeverSpam, f.Actions.NeverSpam)
	e.setProp(propTrash, f.Actions.Trash)
	e.setProp(propNeverImportant, f.Actions.NeverImportant)
	return nil
}

func (e *entry) convertTo(f *filter) error {
	for _, p := range e.Properties {
		switch p.Name {
		// Match
		case propFrom:
			f.Match.From = p.Value
		case propTo:
			f.Match.To = p.Value
		case propSubject:
			f.Match.Subject = p.Value
		case propHasWord:
			if err := f.Match.HasWord.from(p.Value); err != nil {
				return err
			}
		case propNotHaveWord:
			f.Match.NotHaveWord = p.Value

			// Actions
		case propLabel:
			f.Actions.Labels = append(f.Actions.Labels, p.Value)
			sort.Strings(f.Actions.Labels)
		case propMarkAsRead:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Actions.MarkAsRead = true
		case propArchive:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Actions.Archive = true
		case propNeverSpam:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Actions.NeverSpam = true
		case propTrash:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Actions.Trash = true
		case propNeverImportant:
			if p.Value != "true" {
				return errors.New("unexpected value")
			}
			f.Actions.NeverImportant = true

			// Ignored
		case propOperator:
		case propUnit:

		default:
			return fmt.Errorf("unknown property %q", p.Name)
		}
	}
	return nil
}

type property struct {
	Name  propertyName `xml:"name,attr"`
	Value string       `xml:"value,attr"`
}
