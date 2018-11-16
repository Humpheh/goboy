#include <objc/runtime.h>

typedef enum {
	MSG_YESNO,
	MSG_ERROR,
} AlertStyle;

typedef struct {
	char* msg;
	char* title;
	AlertStyle style;
} AlertDlgParams;

typedef struct {
	int save; /* non-zero => save dialog, zero => open dialog */
	char* buf; /* buffer to store selected file */
	int nbuf; /* number of bytes allocated at buf */
	char* title; /* title for dialog box (can be nil) */
	void** exts; /* list of valid extensions (elements actual type is NSString*) */
	int numext; /* number of items in exts */
	int relaxext; /* allow other extensions? */
} FileDlgParams;

typedef enum {
	DLG_OK,
	DLG_CANCEL,
	DLG_URLFAIL,
} DlgResult;

DlgResult alertDlg(AlertDlgParams*);
DlgResult fileDlg(FileDlgParams*);

void* NSStr(void* buf, int len);
void NSRelease(void* obj);
