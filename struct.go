package main

// Here defined the standard input data from mdbook

type Item struct {
	Name        string     `json:"name"`
	Content     string     `json:"content"`
	Number      []int      `json:"number"`
	SubItems    []*Section `json:"sub_items"`
	Path        string     `json:"path"`
	SourcePath  string     `json:"source_path"`
	ParentNames []string   `json:"parent_names"`
}

type Section struct {
	Chapter *Item `json:"Chapter"`
}

type Book struct {
	Sections []*Section `json:"sections"`
}

type BookConfig struct {
	Authors      []string `json:"authors"`
	Language     string   `json:"language"`
	Multilingual bool     `json:"multilingual"`
	Src          string   `json:"src"`
	Title        string   `json:"title"`
}

type GoEPUBConfig struct {
	Command     string `json:"command"`
	CoverImage  string `json:"cover_image"`
	Description string `json:"description"`
}

type OutputConfig struct {
	GoEPUB *GoEPUBConfig `json:"goepub"`
}

type Config struct {
	Book   *BookConfig   `json:"book"`
	Output *OutputConfig `json:"output"`
}

type InputData struct {
	Version     string  `json:"version"`
	Root        string  `json:"root"`
	Book        *Book   `json:"book"`
	Config      *Config `json:"config"`
	Destination string  `json:"destination"`
}
