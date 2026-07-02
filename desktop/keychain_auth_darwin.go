//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework LocalAuthentication

#include <stdlib.h>
#include <string.h>
#include <dispatch/dispatch.h>
#import <Foundation/Foundation.h>
#import <LocalAuthentication/LocalAuthentication.h>

static char* fastgit_authenticate_device_owner(const char* reason_cstr) {
	@autoreleasepool {
		NSString *reason = @"Authenticate to continue";
		if (reason_cstr != NULL) {
			NSString *fromC = [NSString stringWithUTF8String:reason_cstr];
			if (fromC != nil && [fromC length] > 0) {
				reason = fromC;
			}
		}

		LAContext *context = [[LAContext alloc] init];
		NSError *policyError = nil;
		if (![context canEvaluatePolicy:LAPolicyDeviceOwnerAuthentication error:&policyError]) {
			NSString *msg = policyError != nil ? [policyError localizedDescription] : @"Device owner authentication is unavailable";
			return strdup([msg UTF8String]);
		}

		__block BOOL ok = NO;
		__block NSString *errMsg = nil;
		dispatch_semaphore_t sem = dispatch_semaphore_create(0);
		[context evaluatePolicy:LAPolicyDeviceOwnerAuthentication
				localizedReason:reason
						  reply:^(BOOL success, NSError * _Nullable error) {
			ok = success;
			if (!success && error != nil) {
				errMsg = [error localizedDescription];
			}
			dispatch_semaphore_signal(sem);
		}];

		dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);
		if (!ok) {
			if (errMsg == nil || [errMsg length] == 0) {
				errMsg = @"Authentication failed";
			}
			return strdup([errMsg UTF8String]);
		}
		return NULL;
	}
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

func authenticateWithBiometricsOrPasscode(reason string) error {
	cReason := C.CString(reason)
	defer C.free(unsafe.Pointer(cReason))

	cErr := C.fastgit_authenticate_device_owner(cReason)
	if cErr == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(cErr))
	return errors.New(C.GoString(cErr))
}
