package dialog

import (
	"os"
	"path/filepath"
	"github.com/mattn/go-gtk/gtk"
)

func init() {
	gtk.Init(nil)
}

func closeDialog(dlg *gtk.Dialog) {
	dlg.Destroy()
	/* The Destroy call itself isn't enough to remove the dialog from the screen; apparently
	** that happens once the GTK main loop processes some further events. But if we're
	** in a non-GTK app the main loop isn't running, so we empty the event queue before
	** returning from the dialog functions.
	** Not sure how this interacts with an actual GTK app... */
	for gtk.EventsPending() {
		gtk.MainIteration()
	}
}

func (b *MsgBuilder) yesNo() bool {
	dlg := gtk.NewMessageDialog(nil, 0, gtk.MESSAGE_QUESTION, gtk.BUTTONS_YES_NO, "%s", b.Msg)
	dlg.SetTitle(firstOf(b.Dlg.Title, "Confirm?"))
	defer closeDialog(&dlg.Dialog)
	return dlg.Run() == gtk.RESPONSE_YES
}

func (b *MsgBuilder) error() {
	dlg := gtk.NewMessageDialog(nil, 0, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", b.Msg)
	dlg.SetTitle(firstOf(b.Dlg.Title, "Error"))
	defer closeDialog(&dlg.Dialog)
	dlg.Run()
}

func (b *FileBuilder) load() (string, error) {
	return chooseFile("Load", gtk.FILE_CHOOSER_ACTION_OPEN, b)
}

func (b *FileBuilder) save() (string, error) {
	f, err := chooseFile("Save", gtk.FILE_CHOOSER_ACTION_SAVE, b)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(f)
	if !os.IsNotExist(err) && !Message("%s already exists, overwrite?", filepath.Base(f)).yesNo() {
		return "", Cancelled
	}
	return f, nil
}

func chooseFile(title string, action gtk.FileChooserAction, b *FileBuilder) (string, error) {
	dlg := gtk.NewFileChooserDialog(firstOf(b.Dlg.Title, title), nil, action, gtk.STOCK_OK, gtk.RESPONSE_ACCEPT, gtk.STOCK_CANCEL, gtk.RESPONSE_CANCEL)
	for _, filt := range b.Filters {
		filter := gtk.NewFileFilter()
		filter.SetName(filt.Desc)
		for _, ext := range filt.Extensions {
			filter.AddPattern("*." + ext)
		}
		dlg.AddFilter(filter)
	}
	if b.StartDir != "" {
		dlg.SetCurrentFolder(b.StartDir)
	}
	r := dlg.Run()
	defer closeDialog(&dlg.Dialog)
	if r == gtk.RESPONSE_ACCEPT {
		return dlg.GetFilename(), nil
	}
	return "", Cancelled
}
