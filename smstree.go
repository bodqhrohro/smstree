package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gtk"
    "github.com/emersion/go-mbox"
    "mime"
    "net/mail"
    "os"
    "io"
    "bufio"
    "strings"
)

var tree *gtk.TreeView
var treeStore *gtk.TreeStore
var editor *gtk.TextView

var decoder *mime.WordDecoder

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })
    window.Maximize()

    vbox := gtk.NewVBox(true, 0)
    window.Add(vbox)

    tree = gtk.NewTreeView()
    tree.SetHeadersVisible(false)

    treeStore = gtk.NewTreeStore(gtk.TYPE_STRING)
    tree.SetModel(treeStore)
    var headerColumn *gtk.TreeViewColumn = gtk.NewTreeViewColumn()
    tree.AppendColumn(headerColumn)
    var headerRenderer *gtk.CellRendererText = gtk.NewCellRendererText()
    headerColumn.PackStart(headerRenderer, true)
    headerColumn.AddAttribute(headerRenderer, "text", 0)

    vbox.Add(tree)

    editor = gtk.NewTextView()
    vbox.Add(editor)

    return window
}

func decodeHeader(header string) (string, error) {
    scanner := bufio.NewScanner(strings.NewReader(header))
    scanner.Split(bufio.ScanWords)

    var words []string

    for scanner.Scan() {
        decodedWord, err := decoder.Decode(scanner.Text())
        if err != nil {
            return "", err
        }
        words = append(words, decodedWord)
    }

    return strings.Join(words, ""), nil
}

func addEntryFromMBoxMessage(msg *mail.Message) error {
    headers := msg.Header
    subject := headers.Get("Subject")

    subject, err := decodeHeader(subject)
    if err != nil {
        return err
    }

    var rowPtr gtk.TreeIter
    treeStore.Append(&rowPtr, nil)
    treeStore.SetValue(&rowPtr, 0, subject)

    return nil
}

func readFile(filename string) {
    f, err := os.Open(filename)
    if err != nil {
        os.Stderr.WriteString("Can't open the mbox file")
        return
    }
    defer f.Close()

    fileReader := mbox.NewReader(f)

    for {
        messageReader, err := fileReader.NextMessage()

        if err == io.EOF {
            break
        }

        msg, err := mail.ReadMessage(messageReader)
        if err != nil {
            os.Stderr.WriteString("Bad message, skipping")
            continue
        }

        err = addEntryFromMBoxMessage(msg)
        if err != nil {
            os.Stderr.WriteString("Message parse error, skipping")
            continue
        }
    }
}

func main() {
    decoder = new(mime.WordDecoder)

    gtk.Init(nil)
    window := createWindow()

    window.ShowAll()

    readFile(os.Args[1])

    gtk.Main()
}
