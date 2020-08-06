package cmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var markdownURLsToDocsy = func(s string) string {
	s = strings.ReplaceAll(s, ".md", "")
	s = "../" + s
	return s
}

func docsyPrepend(s, docDir string) string {
	title := strings.ReplaceAll(s, docDir, "")
	title = strings.ReplaceAll(title, ".md", "")
	title = strings.ReplaceAll(title, "_", " ")

	return `---
title: "` + title + `"
linkTitle: "` + title + `"
weight: 1
---
`
}

// GenCLIDocsyMarkDown generates docsy-markdown files from cobra commands
func GenCLIDocsyMarkDown(cmd *cobra.Command, docDir, index string) error {
	err := doc.GenMarkdownTreeCustom(cmd, filepath.Join("./", docDir), func(s string) string { return docsyPrepend(s, docDir) }, markdownURLsToDocsy)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join("./", docDir, "_index.md"), []byte(index), 0644)
	if err != nil {
		return err
	}
	return nil
}
