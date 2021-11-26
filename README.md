A GTK+2 editor for MBox files [WIP]
===================================

I couldn't find any good e-mail client or a specialized editor for viewing and editing separate MBox files, so I decided to write it.

The initial purpose, as it comes from the name, is to store backups of SMS conversations in a well-structured format that preserves senders, dates and relations between messages. The relations, unfortunately, should be assigned manually, as SMS messages don't have headers and message relations, and thus it's non-trivial to detect different conversations between the same parties automatically — that's why a manual editor is needed.

Though, it may be used for other purposes too, such as viewing old mailboxes without a need to import them into some e-mail client.

Build
-----

```
go build -o smstree
```

Make sure to install dependencies needed by [go-gtk](https://github.com/mattn/go-gtk) first.

Usage
-----

Currently the program accepts one argument: a MBox file to work with.

Keybindings:

* `Ctrl+Enter`: save current entry to RAM
* `Ctrl+S`: overwrite the MBox file (not implemented yet)
