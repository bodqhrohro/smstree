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

var statusbar *gtk.Statusbar
var errorContext uint


var decoder *mime.WordDecoder

var messageIdIndex map[string]*gtk.TreeIter
var bodyIndex map[string]string


func printError(msg string) {
    statusbar.Push(errorContext, msg)
    os.Stderr.WriteString(msg + "\n")
}

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })
    window.Maximize()

    vbox := gtk.NewVBox(false, 0)
    window.Add(vbox)

    tree = gtk.NewTreeView()


    treeStore = gtk.NewTreeStore(gtk.TYPE_STRING, gtk.TYPE_STRING, gtk.TYPE_STRING, gtk.TYPE_STRING)
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

    var fromColumn *gtk.TreeViewColumn = gtk.NewTreeViewColumn()
    fromColumn.SetResizable(true)
    fromColumn.SetTitle("From")
    tree.AppendColumn(fromColumn)
    var fromRenderer *gtk.CellRendererText = gtk.NewCellRendererText()
    fromColumn.PackEnd(fromRenderer, false)
    fromColumn.AddAttribute(fromRenderer, "text", 2)

    var toColumn *gtk.TreeViewColumn = gtk.NewTreeViewColumn()
    toColumn.SetResizable(true)
    toColumn.SetTitle("To")
    tree.AppendColumn(toColumn)
    var toRenderer *gtk.CellRendererText = gtk.NewCellRendererText()
    toColumn.PackEnd(toRenderer, false)
    toColumn.AddAttribute(toRenderer, "text", 3)


    treeScroll := gtk.NewScrolledWindow(nil, nil)
    treeScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    treeScroll.Add(tree)
    vbox.PackStart(treeScroll, true, true, 0)

    editorScroll := gtk.NewScrolledWindow(nil, nil)
    editorScroll.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
    editor = gtk.NewTextView()
    editorBuffer = editor.GetBuffer()
    editorScroll.Add(editor)
    vbox.PackStart(editorScroll, true, true, 0)


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


    statusbar = gtk.NewStatusbar()
    errorContext = statusbar.GetContextId("error")
    vbox.PackEnd(statusbar, false, false, 0)

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
            printError(err.Error())
        }
    }

    from, err := headers.AddressList("From")
    fromFirst := ""
    if err == nil {
        fromFirst = from[0].Address
    }
    to, err := headers.AddressList("To")
    toFirst := ""
    if err == nil {
        toFirst = to[0].Address
    }

    var rowPtr gtk.TreeIter
    treeStore.Append(&rowPtr, parentPtr)
    treeStore.Set(&rowPtr, subject, dateTimeShort, fromFirst, toFirst)

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
        printError("Can't open the mbox file")
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
            printError("Bad message, skipping")
            continue
        }

        err = addEntryFromMBoxMessage(msg)
        if err != nil {
            printError("Message parse error, skipping")
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
