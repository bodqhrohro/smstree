package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gtk"
)

var tree *gtk.TreeView
var editor *gtk.TextView

func createWindow() *gtk.Window {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.Connect("destroy", func(ctx *glib.CallbackContext) {
        gtk.MainQuit()
    })

    vbox := gtk.NewVBox(true, 0)
    window.Add(vbox)

    tree = gtk.NewTreeView()
    vbox.Add(tree)
    editor = gtk.NewTextView()
    vbox.Add(editor)

    return window
}

func main() {
    gtk.Init(nil)
    window := createWindow()
    window.ShowAll()
    gtk.Main()
}
