package main

import (
	"fmt"
	"github.com/sqweek/dialog"
)

func main() {
	/* Note that spawning a dialog from a non-graphical app like this doesn't
	** quite work properly in OSX. The dialog appears fine, and mouse
	** interaction works but keypresses go straight through the dialog.
	** I'm guessing it has something to do with not having a main loop? *?
	file, err := dialog.File().Title("Save As").Filter("All Files", "*").Save()
	fmt.Println(file)
	fmt.Println("Error:", err)
}
