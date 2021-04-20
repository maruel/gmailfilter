// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const example = `<?xml version='1.0' encoding='UTF-8'?>
<feed xmlns='http://www.w3.org/2005/Atom' xmlns:apps='http://schemas.google.com/apps/2006'>
	<title>Mail Filters</title>
	<id>stuff</id>
	<updated>2006-01-02T15:04:05Z</updated>
	<author>
		<name>Me</name>
		<email>nobody@gmail.com</email>
	</author>
	<entry>
		<category term='filter'></category>
		<title>Mail Filter</title>
		<id>tag:mail.google.com,2008:filter:123</id>
		<updated>2006-01-02T15:04:05Z</updated>
		<content></content>
		<apps:property name='hasTheWord' value='subject:vacation OR subject:ooo OR subject:sick OR (subject:out of office)'/>
		<apps:property name='label' value='vacation'/>
		<apps:property name='shouldMarkAsRead' value='true'/>
		<apps:property name='sizeOperator' value='s_sl'/>
		<apps:property name='sizeUnit' value='s_smb'/>
	</entry>
	<entry>
		<category term='filter'></category>
		<title>Mail Filter</title>
		<id>tag:mail.google.com,2008:filter:124</id>
		<updated>2006-01-02T15:04:05Z</updated>
		<content></content>
		<apps:property name='hasTheWord' value='list:(noise.example.com)'/>
		<apps:property name='label' value='noise'/>
		<apps:property name='shouldArchive' value='true'/>
		<apps:property name='shouldNeverSpam' value='true'/>
	</entry>
	<entry>
		<category term='filter'></category>
		<title>Mail Filter</title>
		<id>tag:mail.google.com,2008:filter:125</id>
		<updated>2006-01-02T15:04:05Z</updated>
		<content></content>
		<apps:property name='from' value='-maruel'/>
		<apps:property name='to' value='-maruel'/>
		<apps:property name='label' value='autre'/>
	</entry>
</feed>
`

func TestCompact(t *testing.T) {
	data := root{}
	if err := xml.Unmarshal([]byte(example), &data); err != nil {
		t.Fatal(err)
	}

	f := gmailFilters{}
	if err := data.convertTo(&f); err != nil {
		t.Fatal(err)
	}

	buf := bytes.Buffer{}
	g := gmailFilters{}
	g.expand(&f)
	if err := g.toCSV(&buf); err != nil {
		t.Fatal(err)
	}
	want := `From,To,Subject,HasWord,NotHaveWord,Labels,MarkAsRead,Archive,NeverSpam,Trash,NeverImportant
-maruel,-maruel,,,,autre,FALSE,FALSE,FALSE,FALSE,FALSE
,,,list:(noise.example.com),,noise,FALSE,TRUE,TRUE,FALSE,FALSE
,,,(subject:out of office),,vacation,TRUE,FALSE,FALSE,FALSE,FALSE
,,,subject:ooo,,vacation,TRUE,FALSE,FALSE,FALSE,FALSE
,,,subject:sick,,vacation,TRUE,FALSE,FALSE,FALSE,FALSE
,,,subject:vacation,,vacation,TRUE,FALSE,FALSE,FALSE,FALSE
`
	if diff := cmp.Diff(buf.String(), want); diff != "" {
		t.Fatal(diff)
	}

	out := root{}
	if err := out.convertFrom(&g); err != nil {
		t.Fatal(err)
	}

	h := gmailFilters{}
	g.compact(&h)
}
