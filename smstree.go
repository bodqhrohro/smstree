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

var messageIndex map[string]*gtk.TreeIter

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

    treeScroll := gtk.NewScrolledWindow(nil, nil)
    treeScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    treeScroll.Add(tree)
    vbox.Add(treeScroll)

    editorScroll := gtk.NewScrolledWindow(nil, nil)
    editorScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    editor = gtk.NewTextView()
    editorScroll.Add(editor)
    vbox.Add(editorScroll)

    return window
}

func decodeHeader(header string) string {
    scanner := bufio.NewScanner(strings.NewReader(header))
    scanner.Split(bufio.ScanWords)

    var words []string

    for scanner.Scan() {
        decodedWord, err := decoder.Decode(scanner.Text())
        if err != nil {
            // probably not encoded, return as is
            return header
        }
        words = append(words, decodedWord)
    }

    return strings.Join(words, "")
}

func addEntryFromMBoxMessage(msg *mail.Message) error {
    headers := msg.Header

    subject := headers.Get("Subject")
    subject = decodeHeader(subject)

    var parentPtr *gtk.TreeIter = nil
    inReplyTo := headers.Get("In-Reply-To")
    if inReplyTo != "" {
        parentPtr = messageIndex[inReplyTo]
    }

    var rowPtr gtk.TreeIter
    treeStore.Append(&rowPtr, parentPtr)
    treeStore.SetValue(&rowPtr, 0, subject)

    messageId := headers.Get("Message-ID")
    if messageId != "" {
        messageIndex[messageId] = &rowPtr
    }

    return nil
}

func readFile(filename string) {
    f, err := os.Open(filename)
    if err != nil {
        os.Stderr.WriteString("Can't open the mbox file\n")
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
            os.Stderr.WriteString("Bad message, skipping\n")
            continue
        }

        err = addEntryFromMBoxMessage(msg)
        if err != nil {
            os.Stderr.WriteString("Message parse error, skipping\n")
            continue
        }
    }
}

func main() {
    decoder = new(mime.WordDecoder)
    messageIndex = make(map[string]*gtk.TreeIter)

    gtk.Init(nil)
    window := createWindow()

    window.ShowAll()

    readFile(os.Args[1])

    gtk.Main()
}
