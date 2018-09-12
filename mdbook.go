package mdbook

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type block string

// Merge looks for files in markdown format and merge them together. It only care following files
// reface.md
// files under folder whose name start with 'ch'.ie. 'ch1'
func Merge(bookFolder string, outFile string) {
	files := bookFiles2Merge(bookFolder)
	mergedFile, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer mergedFile.Close()
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err)
		}
		mergedFile.Write(data)
	}
}

// ReplaceBrokenCode look into the brokenFile, find all block of code and replace them with the block of code in orgFile.
func ReplaceBrokenCode(orgFile, brokenFile, outFile string) {
	codeblocks := codeFromMarkdown(orgFile)
	replaceCodeToMarkdown(codeblocks, brokenFile, outFile)
}

func codeFromMarkdown(file string) []block {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buff := bufio.NewScanner(f)
	begin := false
	blocks := make([]block, 0)
	code := ""
	for buff.Scan() {
		line := buff.Text()
		if !begin && strings.HasPrefix(line, "```") {
			begin = true
			code = line + "\n"
			continue
		}
		if begin && strings.HasPrefix(line, "```") {
			code += "```\n"
			begin = false
			blocks = append(blocks, block(code))
			continue
		}
		if begin {
			code += line + "\n"
		}
	}
	return blocks
}

func replaceCodeToMarkdown(blocks []block, org string, out string) {
	orgFile, err := os.Open(org)
	if err != nil {
		panic(err)
	}
	defer orgFile.Close()
	buff := bufio.NewScanner(orgFile)
	begin := false
	i := 0
	outFile, err := os.Create(out)
	defer outFile.Close()
	if err != nil {
		panic(err)
	}
	for buff.Scan() {
		line := buff.Text()
		if !begin && strings.HasPrefix(line, "```") {
			begin = true
			continue
		}
		if begin && strings.HasPrefix(line, "```") {
			begin = false
			outFile.WriteString(string(blocks[i]))
			i++
			continue
		}
		if !begin {
			outFile.WriteString(line)
			outFile.WriteString("\n")
		}
	}
}

func bookFiles2Merge(bookFolder string) []string {
	files2Merge := make([]string, 0)
	folders, err := ioutil.ReadDir(bookFolder)
	if err != nil {
		panic("cannot read book dir: " + err.Error())
	}
	sort.Slice(folders, func(i, j int) bool {
		v := strings.Compare(folders[i].Name(), folders[j].Name())
		if v == 1 {
			return false
		}
		return true
	})
	chaps := make([]os.FileInfo, 0)
	for _, folder := range folders {
		if strings.HasPrefix(folder.Name(), "ch") {
			chaps = append(chaps, folder)
		}
	}
	for _, chap := range chaps {
		files, err := ioutil.ReadDir(filepath.Join(bookFolder, chap.Name()))
		if err != nil {
			log.Fatal(err)
		}
		sort.Slice(files, func(i, j int) bool {
			if files[i].Name() == "readme.md" {
				return true
			}
			if files[j].Name() == "readme.md" {
				return false
			}
			v := strings.Compare(files[i].Name(), files[j].Name())
			if v == 1 {
				return false
			}
			return true
		})
		for _, file := range files {
			files2Merge = append(files2Merge, filepath.Join(bookFolder, chap.Name(), file.Name()))
		}
	}
	return files2Merge
}
