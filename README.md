# gomdbook2epub

An EPUB generator backend for [mdBook](https://github.com/rust-lang/mdBook).

## Getting Started

1. Install binary from [release page](https://github.com/mudream4869/gomdbook2epub/releases)

2. Add config to book.toml

```toml
[output.goepub]
description = "[Description]"
cover_image = "[path to book cover image]"
command = "[path to gomdbook2epub binary]"
```

## TODO

- [ ] Add default css files for cover and book.
- [ ] Allow user to config custom css files.

## Known Issue

Currently, go-epub only allow us to add content in the highest level.
