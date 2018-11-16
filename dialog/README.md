# dialog
Simple cross-platform dialog API for go-lang

# examples
    ok := dialog.Message("%s", "Do you want to continue?").Title("Are you sure?").YesNo()

Creates a dialog box titled "Are you sure?", containing the message "Do you want to continue?",
a "Yes" button and a "No" button. Returns true iff the dialog could be displayed and the user
pressed the "Yes" button.

    filename, err := dialog.File().Filter("Mp3 audio file", "mp3").Load()

Creates a file selection dialog allowing the user to select a .mp3 file. The absolute path of
the file is returned, unless an error is encountered or the user cancels/closes the dialog.
In the latter case, `filename` will be the empty string and `err` will equal `dialog.Cancelled`.

    filename, err := dialog.File().Filter("XML files", "xml").Title("Export to XML").Save()

Asks the user for a filename to write data into. If the user selects a file which already exists,
an additional dialog is spawned to confirm they want to overwrite the existing file.

# platform details
* OSX: uses Cocoa's NSAlert/NSSavePanel/SOpenPanel clasess
* Win32: uses MessageBox/GetOpenFileName/GetSaveFileName (via package github.com/AllenDang/w32)
* Linux: uses Gtk's MessageDialog/FileChooserDialog (via package github.com/mattn/gtk)
