package dialog

import (
	"fmt"
	"reflect"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/TheTitanrain/w32"
)

type WinDlgError int

func (e WinDlgError) Error() string {
	return fmt.Sprintf("CommDlgExtendedError: %#x", e)
}

func err() error {
	e := w32.CommDlgExtendedError()
	if e == 0 {
		return Cancelled
	}
	return WinDlgError(e)
}

func (b *MsgBuilder) yesNo() bool {
	r := w32.MessageBox(w32.HWND(0), b.Msg, firstOf(b.Dlg.Title, "Confirm?"), w32.MB_YESNO)
	return r == w32.IDYES
}

func (b *MsgBuilder) error() {
	w32.MessageBox(w32.HWND(0), b.Msg, firstOf(b.Dlg.Title, "Error"), w32.MB_OK|w32.MB_ICONERROR)
}

type filedlg struct {
	buf     []uint16
	filters []uint16
	opf     *w32.OPENFILENAME
}

func (d filedlg) Filename() string {
	i := 0
	for i < len(d.buf) && d.buf[i] != 0 {
		i++
	}
	return string(utf16.Decode(d.buf[:i]))
}

func (b *FileBuilder) load() (string, error) {
	d := openfile(w32.OFN_FILEMUSTEXIST, b)
	if w32.GetOpenFileName(d.opf) {
		return d.Filename(), nil
	}
	return "", err()
}

func (b *FileBuilder) save() (string, error) {
	d := openfile(w32.OFN_OVERWRITEPROMPT, b)
	if w32.GetSaveFileName(d.opf) {
		return d.Filename(), nil
	}
	return "", err()
}

/* syscall.UTF16PtrFromString not sufficient because we need to encode embedded NUL bytes */
func utf16ptr(utf16 []uint16) *uint16 {
	if utf16[len(utf16)-1] != 0 {
		panic("refusing to make ptr to non-NUL terminated utf16 slice")
	}
	h := (*reflect.SliceHeader)(unsafe.Pointer(&utf16))
	return (*uint16)(unsafe.Pointer(h.Data))
}

func utf16slice(ptr *uint16) []uint16 {
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(ptr)), Len: 1, Cap: 1}
	slice := *((*[]uint16)(unsafe.Pointer(&hdr)))
	i := 0
	for slice[len(slice)-1] != 0 {
		i++
	}
	hdr.Len = i
	slice = *((*[]uint16)(unsafe.Pointer(&hdr)))
	return slice
}

func openfile(flags uint32, b *FileBuilder) (d filedlg) {
	d.buf = make([]uint16, w32.MAX_PATH)
	d.opf = &w32.OPENFILENAME{
		File:    utf16ptr(d.buf),
		MaxFile: uint32(len(d.buf)),
		Flags:   flags,
	}
	d.opf.StructSize = uint32(unsafe.Sizeof(*d.opf))
	if b.StartDir != "" {
		d.opf.InitialDir, _ = syscall.UTF16PtrFromString(b.StartDir)
	}
	if b.Dlg.Title != "" {
		d.opf.Title, _ = syscall.UTF16PtrFromString(b.Dlg.Title)
	}
	for _, filt := range b.Filters {
		/* build utf16 string of form "Music File\0*.mp3;*.ogg;*.wav;\0" */
		d.filters = append(d.filters, utf16.Encode([]rune(filt.Desc))...)
		d.filters = append(d.filters, 0)
		for _, ext := range filt.Extensions {
			s := fmt.Sprintf("*.%s;", ext)
			d.filters = append(d.filters, utf16.Encode([]rune(s))...)
		}
		d.filters = append(d.filters, 0)
	}
	if d.filters != nil {
		d.filters = append(d.filters, 0, 0) // two extra NUL chars to terminate the list
		d.opf.Filter = utf16ptr(d.filters)
	}
	return d
}
