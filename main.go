package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/bbalet/stopwords"
	"github.com/reiver/go-porterstemmer"
)

// A document, which is a file and some info about its words and clustering.
// Also used for clustering -- we repurpose the same struct type.
type Doc struct {
	path      string
	file      os.FileInfo
	wordCount map[string]float64
	clusterId int // -1 if it's a doc, 0..k if it's a cluster
	size      float64
}

func main() {
	docs := make([]*Doc, 0)
	allWords := make(map[string]bool)
	clusters := 50
	minWord := 5
	pattern := ".md"

	// Recursively find all matching files inside the directory and process them.
	flag.Parse()
	root := flag.Arg(0)
	err := filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() && strings.HasSuffix(f.Name(), pattern) {
				doc := &Doc{
					path:      path,
					file:      f,
					clusterId: -1, // will be used to check for convergence
					wordCount: make(map[string]float64),
				}
				err := countWords(doc, allWords, minWord)
				if err != nil {
					fmt.Println(err)
				} else {
					docs = append(docs, doc)
				}
			}
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}

	// Perform k-means clustering on the array of docs
	results, err := kmeans(clusters, docs, allWords)
	if err != nil {
		fmt.Println(err)
		return
	}

	topWords := make([]string, 0, len(docs))
	for word := range allWords {
		topWords = append(topWords, word)
	}

	// Print out each cluster, and its most common words; then print docs
	// that belong to that cluster.
	fmt.Printf("Clustered %d docs with %d words into %d clusters\n",
		len(docs), len(allWords), len(results))
	for i, clust := range results {
		fmt.Printf("Cluster %d, %.f documents\n", i, clust.size)
		fmt.Println("    Top words:")
		sort.Slice(topWords, func(i, j int) bool {
			return clust.wordCount[topWords[i]] > clust.wordCount[topWords[j]]
		})
		for i := 0; i < 15; i++ {
			fmt.Printf("    %s\t%.2f\n", topWords[i], clust.wordCount[topWords[i]])
		}
		fmt.Println("    Documents:")
		limit := 20
		for _, doc := range docs {
			if limit > 0 && doc.clusterId == i {
				limit--
				fmt.Printf("    %s\n", doc.path)
			}
		}
	}

}

// Perform k-means clustering. Cluster the documents into k clusters (must be <
// number of documents). Clusters are initialized using the random method, by
// choosing some of the documents to be the initial means (centroids). We
// are done after N iterations, or after no docs get clusterIDs reassigned
func kmeans(k int, docs []*Doc, allWords map[string]bool) ([]*Doc, error) {
	if len(docs) <= 2 || len(docs) <= k {
		return nil, fmt.Errorf("Not enough documents to cluster: %d", len(docs))
	}
	clusters := make([]*Doc, 0, k) // These are the "means" or "centroids"
	iters := 20
	converged := false

	// Pick k docs at random, without replacement
	chosen := map[int]bool{}
	for len(clusters) < k {
		i := rand.Intn(len(docs))
		if !chosen[i] {
			chosen[i] = true
			doc := &Doc{
				wordCount: make(map[string]float64),
			}
			for word, count := range docs[i].wordCount {
				doc.wordCount[word] = count
			}
			clusters = append(clusters, doc)
		}
	}

	for iters > 0 && !converged {
		iters--
		converged = true

		// For each doc, find the closest cluster centroid and assign
		// the document to that cluster
		for _, doc := range docs {
			min := math.MaxFloat64
			for i, clust := range clusters {
				dist := 0.0
				for word := range allWords {
					dist += math.Pow(doc.wordCount[word]-clust.wordCount[word], 2)
				}
				if dist < min {
					min = dist
					if doc.clusterId != i {
						converged = false
					}
					doc.clusterId = i
				}
			}
		}

		// "Average" the documents within each cluster to generate new means
		if !converged {
			for i, clust := range clusters {
				clust.wordCount = make(map[string]float64) // reset wordcounts to zero
				clust.size = 0
				for _, doc := range docs {
					if doc.clusterId == i {
						for word := range allWords {
							clust.wordCount[word] += doc.wordCount[word]
						}
						clust.size++
					}
				}
				if clust.size > 0 {
					for word := range allWords {
						clust.wordCount[word] /= clust.size
					}
				}
			}
		}
	}

	return clusters, nil
}

// Read a file and count the stemmed, stopworded words in it.
func countWords(doc *Doc, allWords map[string]bool, minLength int) error {
	// Get the file's contents as a string
	b, err := ioutil.ReadFile(doc.path)
	if err != nil {
		return err
	}

	// Remove stop words.
	str := stopwords.CleanString(string(b), "en", true)

	// Split the string into words, removing punctuation
	noPunctuation := func(r rune) rune {
		if strings.ContainsRune("'’", r) {
			return -1
		} else {
			return r
		}
	}
	str = strings.Map(noPunctuation, str)
	words := strings.FieldsFunc(str, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	// Stem the words, increment their counts, and track global word existence
	// across all files.
	for _, word := range words {
		word = porterstemmer.StemString(strings.ToLower(word))
		if len(word) >= minLength {
			allWords[word] = true
			if c, ok := doc.wordCount[word]; !ok {
				doc.wordCount[word] = 1
			} else {
				doc.wordCount[word] = c + 1
			}
		}
	}

	return nil
}
