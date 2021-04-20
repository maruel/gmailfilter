// Copyright 2021 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// gmailfilter helps manage GMail filters.
package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
)

/*
func prettyPrint(data interface{}) error {
	//fmt.Printf("%+v\n", data)
	var p []byte
	p, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Printf("%s\n", p)
	return err
}
*/

func mainImpl() error {
	if len(os.Args) != 2 {
		return errors.New("usage: gmailfilter xml")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	data := root{}
	if err = xml.Unmarshal(raw, &data); err != nil {
		return err
	}
	//prettyPrint(data)

	f := gmailFilters{}
	if err = data.convertTo(&f); err != nil {
		return err
	}
	//prettyPrint(f)

	g := gmailFilters{}
	g.expand(&f)
	if err = g.toCSV(os.Stdout); err != nil {
		return err
	}

	out := root{}
	if err = out.convertFrom(&g); err != nil {
		return err
	}
	//prettyPrint(out)
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "gmailfilter: %s.\n", err)
		os.Exit(1)
	}
}
