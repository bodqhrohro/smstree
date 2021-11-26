package main

/*
#include <glib.h>

extern gboolean goSaveCallback();
extern gboolean goSaveRecordCallback();

gboolean saveCallbackCgo() {
    return goSaveCallback();
}

gboolean saveRecordCallbackCgo() {
    return goSaveRecordCallback();
}
*/
import "C"
