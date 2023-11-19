// SPDX-License-Identifier: Unlicense OR MIT

#import <Foundation/Foundation.h>

#include "_cgo_export.h"

void gio_wakeupMainThread(void) {
	dispatch_async(dispatch_get_main_queue(), ^{
		gio_dispatchMainFuncs();
	});
}
