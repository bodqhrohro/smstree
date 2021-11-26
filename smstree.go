// +build !cgocheck

package main

/*
#include "smstree.go.h"
#cgo pkg-config: glib-2.0 gobject-2.0 gtk+-2.0

extern gboolean saveCallbackCgo();
extern gboolean saveRecordCallbackCgo();
*/
import "C"

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gdk"
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
    "unsafe"
)

var window *gtk.Window
var filename string

var tree *gtk.TreeView
var treeStore *gtk.TreeStore
var editor *gtk.TextView
var editorBuffer *gtk.TextBuffer
var treeSelection *gtk.TreeSelection

var statusbar *gtk.Statusbar
var errorContext uint

var currentPath string


var decoder *mime.WordDecoder

var messageIdIndex map[string]*gtk.TreeIter
var bodyIndex map[string]string

var dirtyMessage bool
var suppressNextEditorChanged bool
// var suppressNextTreeChanged bool


func printError(msg string) {
    statusbar.Push(errorContext, msg)
    os.Stderr.WriteString(msg + "\n")
}

func getCurrentTreePath() string {
    var selectedRowPtr gtk.TreeIter
    treeSelection.GetSelected(&selectedRowPtr)
    return treeStore.GetPath(&selectedRowPtr).String()
}

func setDirtyMessage(dirty bool) {
    if (dirty) {
        window.SetTitle("*" + formatTitle())
        tree.SetSensitive(false)
    } else {
        window.SetTitle(formatTitle())
        tree.SetSensitive(true)
    }
    dirtyMessage = dirty
}

func getBufferText(b *gtk.TextBuffer) string {
    var startIter, endIter gtk.TextIter
    b.GetStartIter(&startIter)
    b.GetEndIter(&endIter)
    return b.GetText(&startIter, &endIter, true)
}

func saveMessage() {
    bodyIndex[currentPath] = getBufferText(editorBuffer)

    setDirtyMessage(false)
}

func confirm(text string) bool {
    dialog := gtk.NewDialog()
    dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL)
    dialog.AddButton("OK", gtk.RESPONSE_OK)
    dialog.SetDestroyWithParent(true)
    dialog.SetTitle(text)
    dialog.SetDefaultSize(500, 30)

    /* label := gtk.NewLabel(text)
    contentArea := dialog.GetContentArea()
    contentArea.PackStart(label, true, true, 0)
    contentArea.Show() */

    response := dialog.Run()
    dialog.Destroy()

    if response == gtk.RESPONSE_OK {
        return true
    } else {
        return false
    }
}

func formatTitle() string {
    return filename + " — smstree"
}

func addAccel(window *gtk.Window, key C.uint, modifier gdk.ModifierType, cb unsafe.Pointer) {
    accelGroup := gtk.NewAccelGroup()
    C.gtk_accel_group_connect(
        (*C.GtkAccelGroup)(unsafe.Pointer(accelGroup.GAccelGroup)),
        key,
        C.GdkModifierType(modifier),
        C.GtkAccelFlags(0),
        C.g_cclosure_new(C.GCallback(cb), nil, nil),
    )
    window.AddAccelGroup(accelGroup);
}

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })
    window.Maximize()
    window.SetTitle(formatTitle())

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
        /* if suppressNextTreeChanged {
            suppressNextTreeChanged = false
            return
        }

        if dirtyMessage {
            if confirm("The message was modified. Save?") {
                saveMessage()
            } else {
                suppressNextTreeChanged = true

                go func() {
                    path := gtk.NewTreePathFromString(currentPath)
                    tree.SetCursor(path.Copy(), nil, false)
                    editor.GrabFocus()
                }()

                return
            }
        } */
        currentPath = getCurrentTreePath()
        body, ok := bodyIndex[currentPath]
        suppressNextEditorChanged = true
        if ok {
            editorBuffer.SetText(body)
        } else {
            editorBuffer.SetText("")
        }
    })

    editorBuffer.Connect("changed", func() {
        if suppressNextEditorChanged {
            if (getBufferText(editorBuffer) != "") {
                suppressNextEditorChanged = false
            }
        } else {
            setDirtyMessage(true)
        }
    })

    addAccel(window, gdk.KEY_s, gdk.CONTROL_MASK, C.saveCallbackCgo)
    addAccel(window, gdk.KEY_Return, gdk.CONTROL_MASK, C.saveRecordCallbackCgo)


    statusbar = gtk.NewStatusbar()
    errorContext = statusbar.GetContextId("error")
    vbox.PackEnd(statusbar, false, false, 0)

    return window
}

//export goSaveCallback
func goSaveCallback() C.gboolean {
    saveMessage()

    return C.gboolean(1);
}

//export goSaveRecordCallback
func goSaveRecordCallback() C.gboolean {
    saveMessage()

    return C.gboolean(1);
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
    filename = os.Args[1]

    decoder = new(mime.WordDecoder)
    messageIdIndex = make(map[string]*gtk.TreeIter)
    bodyIndex = make(map[string]string)

    gtk.Init(nil)
    window = createWindow()

    window.ShowAll()

    readFile(filename)

    gtk.Main()
}
