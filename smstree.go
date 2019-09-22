package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gtk"
    "github.com/emersion/go-mbox"
    "mime"
    "net/mail"
    "os"
    "io"
    "io/ioutil"
    "bufio"
    "strings"
    "time"
)

var tree *gtk.TreeView
var treeStore *gtk.TreeStore
var editor *gtk.TextView
var editorBuffer *gtk.TextBuffer
var treeSelection *gtk.TreeSelection

var decoder *mime.WordDecoder

var messageIdIndex map[string]*gtk.TreeIter
var bodyIndex map[string]string

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })
    window.Maximize()

    vbox := gtk.NewVBox(true, 0)
    window.Add(vbox)

    tree = gtk.NewTreeView()


    treeStore = gtk.NewTreeStore(gtk.TYPE_STRING, gtk.TYPE_STRING)
    tree.SetModel(treeStore)

    var headerColumn *gtk.TreeViewColumn = gtk.NewTreeViewColumn()
    headerColumn.SetResizable(true)
    tree.AppendColumn(headerColumn)
    var headerRenderer *gtk.CellRendererText = gtk.NewCellRendererText()
    headerColumn.PackStart(headerRenderer, true)
    headerColumn.AddAttribute(headerRenderer, "text", 0)

    var dateColumn *gtk.TreeViewColumn = gtk.NewTreeViewColumn()
    dateColumn.SetTitle("Datetime")
    tree.AppendColumn(dateColumn)
    var dateRenderer *gtk.CellRendererText = gtk.NewCellRendererText()
    dateColumn.PackEnd(dateRenderer, false)
    dateColumn.AddAttribute(dateRenderer, "text", 1)


    treeScroll := gtk.NewScrolledWindow(nil, nil)
    treeScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    treeScroll.Add(tree)
    vbox.Add(treeScroll)

    editorScroll := gtk.NewScrolledWindow(nil, nil)
    editorScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    editor = gtk.NewTextView()
    editorBuffer = editor.GetBuffer()
    editorScroll.Add(editor)
    vbox.Add(editorScroll)


    treeSelection = tree.GetSelection()
    treeSelection.Connect("changed", func() {
        var selectedRowPtr gtk.TreeIter
        treeSelection.GetSelected(&selectedRowPtr)
        body, ok := bodyIndex[treeStore.GetPath(&selectedRowPtr).String()]
        if ok {
            editorBuffer.SetText(body)
        } else {
            editorBuffer.SetText("")
        }
    })


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
        parentPtr = messageIdIndex[inReplyTo]
    }

    dateTimeRFC := headers.Get("Date")
    dateTimeShort := ""
    if dateTimeRFC != "" {
        dateTime, err := time.Parse("Mon, _2 Jan 2006 15:04:05 -0700", dateTimeRFC)
        if err == nil {
            dateTimeShort = dateTime.Format("02/01/2006 15:04:05")
        } else {
            os.Stderr.WriteString(err.Error() + "\n")
        }
    }

    var rowPtr gtk.TreeIter
    treeStore.Append(&rowPtr, parentPtr)
    treeStore.Set(&rowPtr, subject, dateTimeShort)

    messageId := headers.Get("Message-ID")
    if messageId != "" {
        messageIdIndex[messageId] = &rowPtr
    }

    body, err := ioutil.ReadAll(msg.Body)
    if err == nil {
        body := string(body)
        bodyIndex[treeStore.GetPath(&rowPtr).String()] = body
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
    messageIdIndex = make(map[string]*gtk.TreeIter)
    bodyIndex = make(map[string]string)

    gtk.Init(nil)
    window := createWindow()

    window.ShowAll()

    readFile(os.Args[1])

    gtk.Main()
}
