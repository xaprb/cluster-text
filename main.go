package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type FileWords struct {
	path      string
	file      os.FileInfo
	wordCount map[string]int
}

func main() {
	wordCounts := make([]FileWords, 0)
	allWords := make(map[string]bool)

	// Recursively find all *.md files inside the directory and process them.
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".md") {
				fw := FileWords{
					path:      path,
					file:      f,
					wordCount: make(map[string]int),
				}
				err := countWords(fw, allWords)
				if err != nil {
					fmt.Println(err)
				} else {
					wordCounts = append(wordCounts, fw)
				}
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}

	// Print out each file and its words and their counts.
	for _, fw := range wordCounts {
		fmt.Println(fw.path)
		for word, c := range fw.wordCount {
			fmt.Println("   ", word, c)
		}
	}

}

// Read a file and count the words in it.
func countWords(fw FileWords, allWords map[string]bool) error {
	// Get the file's contents as a string
	b, err := ioutil.ReadFile(fw.path)
	if err != nil {
		return err
	}

	// Split the string into words, removing punctuation
	noPunctuation := func(r rune) rune {
		if strings.ContainsRune("'’", r) {
			return -1
		} else {
			return r
		}
	}
	str := strings.Map(noPunctuation, string(b))
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	// Iterate the words. For each one, insert it into the map of all words, and
	// increment its count.
	for _, word := range words {
		word = strings.ToLower(word)
		if len(word) > 3 {
			// stem the words
			// remove stopwords
			allWords[word] = true
			if c, ok := fw.wordCount[word]; !ok {
				fw.wordCount[word] = 1
			} else {
				fw.wordCount[word] = c + 1
			}
		}
	}

	return nil
}
