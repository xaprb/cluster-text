# cluster-text
Document clustering with basic machine learning

This is a pretty naive "script" in Go, that looks for `*.md` files in a directory or its subdirectories, reads them, parses them into words, stems and eliminates stopwords, and then does K-Means Clustering to categorize the texts.

It's a Sunday project to see if I could get auto-categorization for my blog posts.

It's not super-fast, and my guess is because it uses a bunch of maps of strings instead of arrays:

    cluster-text $ time go run main.go ~/repos/xaprb/xaprb-src/content/blog > output.txt
    real	26m18.683s
    user	26m13.341s
    sys	0m8.029s
    
    Clustered 1158 docs with 9877 words into 50 clusters
    
