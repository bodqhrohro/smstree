package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gtk"
)

var tree *gtk.TreeView
var treeStore *gtk.TreeStore
var editor *gtk.TextView

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })

    vbox := gtk.NewVBox(true, 0)
    window.Add(vbox)

    tree = gtk.NewTreeView()
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

func main() {
    gtk.Init(nil)
    window := createWindow()

    var rowPtr gtk.TreeIter
    treeStore.Append(&rowPtr, nil)
    treeStore.SetValue(&rowPtr, 0, "a")

    var subRowPtr gtk.TreeIter
    treeStore.Append(&subRowPtr, &rowPtr)
    treeStore.SetValue(&subRowPtr, 0, "b")

    window.ShowAll()
    gtk.Main()
}
