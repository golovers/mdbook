package main

import "github.com/golovers/mdbook"

func main() {
	bookFolder := "C:/Users/pthethanh/go/src/github.com/chai2010/advanced-go-programming-book/"
	mdbook.Merge(bookFolder, "merged.md")

	// chinse-good-code contains original Chinese lang and good block of codes
	// english-broken-code is translated into English but the blocks of code are broken
	mdbook.ReplaceBrokenCode("chinese-good-code.md", "english-broken-code.md", "english-good-code.md")
}
