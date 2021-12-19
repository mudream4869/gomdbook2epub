package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	epub "github.com/bmaupin/go-epub"

	"github.com/mudream4869/gomdbook2epub/srcreplacer"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	imgSourceMap map[string]string
}

func newRenderer(book *Book) *Renderer {
	return &Renderer{
		imgSourceMap: map[string]string{},
	}
}

func (r *Renderer) hashFilename(src string) string {
	h := sha1.New()
	h.Write([]byte(src))
	return base64.URLEncoding.EncodeToString(h.Sum(nil)) + "-" + path.Base(src)
}

func (r *Renderer) chapterNumbersToString(numbers []int) string {
	var ret string
	for _, n := range numbers {
		ret += fmt.Sprintf("%d.", n)
	}
	return ret
}

func (r *Renderer) renderItem(item *Item, absRoot string, ebook *epub.Epub) error {
	if len(item.SourcePath) == 0 {
		// Skip the draft
		return nil
	}

	absFolderPath := path.Join(absRoot, filepath.Dir(item.SourcePath))
	absFilePath := path.Join(absRoot, item.SourcePath)

	imgSrcFunc := func(src string) string {
		// Register image src if source is from local

		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			// We don't need to add image from web, goepub will handle it.
			return src
		}

		if !path.IsAbs(src) {
			src = path.Clean(path.Join(absFolderPath, src))
		}

		target, exist := r.imgSourceMap[src]
		if exist {
			return target
		}

		target = r.hashFilename(src)
		log.Println("Renderer.run: Add Image:", src, "to", target)

		internalImgPath, err := ebook.AddImage(src, target)
		if err != nil {
			log.Println("Renderer.run: [Warning] Adding image not succeceed:", err)
			return target
		}

		r.imgSourceMap[src] = internalImgPath
		return internalImgPath
	}

	linkSrcFunc := func(src string) string {
		// Change link src if source is a local file

		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
			return src
		}

		hashTagPos := strings.IndexRune(src, '#')
		if hashTagPos == 0 {
			return src
		}

		filename := src
		hashTag := ""
		if hashTagPos != -1 {
			filename = src[:hashTagPos]
			hashTag = src[hashTagPos:]
		}

		if !strings.HasSuffix(filename, ".md") {
			log.Println("Renderer.run: [Warning] link to a none markdown local file:", src)
			return src
		}

		if !path.IsAbs(filename) {
			filename = path.Clean(path.Join(absFolderPath, filename))
		}

		if !strings.HasPrefix(filename, absFolderPath) {
			log.Println("Renderer.run: [Warning] file not in mdbook src folder: ", filename)
		}

		return r.hashFilename(filename) + ".xhtml" + hashTag
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	err := md.Convert([]byte(item.Content), &buf)
	if err != nil {
		return fmt.Errorf("Renderer.renderItem: %w", err)
	}

	html, err := srcreplacer.ReplaceHTML(buf.Bytes(), linkSrcFunc, imgSrcFunc)
	if err != nil {
		return fmt.Errorf("Renderer.renderItem: %w", err)
	}

	title := fmt.Sprintf("%s %s", r.chapterNumbersToString(item.Number), item.Name)
	internalFilename := r.hashFilename(absFilePath) + ".xhtml"

	log.Println("Renderer.renderItem: Add", absFilePath, "to", internalFilename)
	ebook.AddSection(string(html), title, internalFilename, "")

	for _, subitem := range item.SubItems {
		err := r.renderItem(subitem.Chapter, absRoot, ebook)
		if err != nil {
			return fmt.Errorf("Renderer.renderItem: %w", err)
		}
	}

	return nil
}

func (r *Renderer) run(inputData *InputData) (*epub.Epub, error) {
	ebook := epub.NewEpub(inputData.Config.Book.Title)
	ebook.SetLang(inputData.Config.Book.Language)
	ebook.SetDescription(inputData.Config.Output.GoEPUB.Description)

	authors := inputData.Config.Book.Authors
	if len(authors) > 0 {
		ebook.SetAuthor(authors[0])
	}

	coverImg := inputData.Config.Output.GoEPUB.CoverImage
	if len(coverImg) > 0 {
		coverImg = path.Join(inputData.Root, coverImg)
		internalImgPath, err := ebook.AddImage(coverImg, "image.png")
		if err != nil {
			return nil, fmt.Errorf("Renderer.run: %w", err)
		}
		ebook.SetCover(internalImgPath, "")
	}

	absRoot := path.Join(inputData.Root, inputData.Config.Book.Src)
	for _, section := range inputData.Book.Sections {
		err := r.renderItem(section.Chapter, absRoot, ebook)
		if err != nil {
			return nil, fmt.Errorf("Renderer.run: %w", err)
		}
	}

	return ebook, nil
}

func Main() error {
	log.Println("Main: loading data from stdin...")
	bs, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("Main: %w", err)
	}

	var inputData InputData
	err = json.Unmarshal(bs, &inputData)
	if err != nil {
		return fmt.Errorf("Main: %w", err)
	}
	log.Println("Main: loading data from stdin...DONE")

	log.Println("Main: rendering...")
	renderer := newRenderer(inputData.Book)
	ebook, err := renderer.run(&inputData)
	if err != nil {
		return fmt.Errorf("Main: %w", err)
	}
	log.Println("Main: rendering...DONE")

	log.Println("Main: writing...")
	outputBookName := path.Join(inputData.Destination, "output.epub")
	err = ebook.Write(outputBookName)
	if err != nil {
		return fmt.Errorf("Main: %w", err)
	}
	log.Println("Main: writing...DONE")

	return nil
}
