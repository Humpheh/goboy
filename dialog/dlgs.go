/* Package dialog provides a simple cross-platform common dialog API.
Eg. to prompt the user with a yes/no dialog:

     if dialog.MsgDlg("%s", "Do you want to continue?").YesNo() {
         // user pressed Yes
     }

The general usage pattern is to call one of the toplevel *Dlg functions
which return a *Builder structure. From here you can optionally call
configuration functions (eg. Title) to customise the dialog, before
using a launcher function to run the dialog.
*/
package dialog

import (
	"errors"
	"fmt"
)

/* Cancelled is an error returned when a user cancels/closes a dialog. */
var Cancelled = errors.New("Cancelled")

type Dlg struct {
	Title string
}

type MsgBuilder struct {
	Dlg
	Msg string
}

/* Message initialises a MsgBuilder with the provided message */
func Message(format string, args ...interface{}) *MsgBuilder {
	return &MsgBuilder{Msg: fmt.Sprintf(format, args...)}
}

/* Title specifies what the title of the message dialog will be */
func (b *MsgBuilder) Title(title string) *MsgBuilder {
	b.Dlg.Title = title
	return b
}

/* YesNo spawns the message dialog with two buttons, "Yes" and "No".
Returns true iff the user selected "Yes". */
func (b *MsgBuilder) YesNo() bool {
	return b.yesNo()
}

/* Error spawns the message dialog with an error icon and single button, "Ok". */
func (b *MsgBuilder) Error() {
	b.error()
}

/* FileFilter represents a category of files (eg. audio files, spreadsheets). */
type FileFilter struct {
	Desc string
	Extensions []string
}

type FileBuilder struct {
	Dlg
	StartDir string
	Filters []FileFilter
}

/* File initialises a FileBuilder using the default configuration. */
func File() *FileBuilder {
	return &FileBuilder{}
}

/* Title specifies the title to be used for the dialog. */
func (b *FileBuilder) Title(title string) *FileBuilder {
	b.Dlg.Title = title
	return b
}

/* Filter adds a category of files to the types allowed by the dialog. Multiple
calls to Filter are cumulative - any of the provided categories will be allowed.
By default all files can be selected.

The special extension '*' allows all files to be selected when the Filter is active. */
func (b *FileBuilder) Filter(desc string, extensions ...string) *FileBuilder {
	filt := FileFilter{desc, extensions}
	if len(filt.Extensions) == 0 {
		filt.Extensions = append(filt.Extensions, "*")
	}
	b.Filters = append(b.Filters, filt)
	return b
}

/* SetStartDir specifies the initial directory of the dialog. */
func (b *FileBuilder) SetStartDir(startDir string) *FileBuilder {
	b.StartDir = startDir
	return b
}

/* Load spawns the file selection dialog using the configured settings,
asking the user to select a single file. Returns Cancelled as the error
if the user cancels or closes the dialog. */
func (b *FileBuilder) Load() (string, error) {
	return b.load()
}

/* Save spawns the file selection dialog using the configured settings,
asking the user for a filename to save as. If the chosen file exists, the
user is prompted whether they want to overwrite the file. Returns
Cancelled as the error if the user cancels/closes the dialog, or selects
not to overwrite the file. */
func (b *FileBuilder) Save() (string, error) {
	return b.save()
}

