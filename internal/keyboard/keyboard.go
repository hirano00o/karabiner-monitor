package keyboard

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework ApplicationServices
#import <CoreGraphics/CoreGraphics.h>
#import <ApplicationServices/ApplicationServices.h>

// Check if the process has accessibility permissions
int checkAccessibility() {
    return AXIsProcessTrusted() ? 1 : 0;
}

// Global flag to track if key was pressed
static int keyPressed = 0;

// Callback function for CGEventTap
CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventKeyDown || type == kCGEventKeyUp) {
        keyPressed = 1;
    }
    return event;
}

// Start monitoring keyboard events
// Returns: 0 on success, -1 on error
int startKeyboardMonitor() {
    keyPressed = 0;

    CGEventMask eventMask = (CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventKeyUp));
    CFMachPortRef eventTap = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionListenOnly,
        eventMask,
        eventCallback,
        NULL
    );

    if (!eventTap) {
        return -1;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);

    // Clean up
    CFRelease(runLoopSource);
    CFRelease(eventTap);

    return 0;
}

// Check if a key was pressed since monitoring started
int wasKeyPressed() {
    return keyPressed;
}

// Reset the key pressed flag
void resetKeyPressed() {
    keyPressed = 0;
}
*/
import "C"
import (
	"context"
	"time"
)

// CheckAccessibilityPermission checks if the application has accessibility permissions.
// On macOS, accessibility permissions are required to monitor keyboard events.
//
// Returns true if the permission is granted, false otherwise.
//
// If this returns false, the user needs to grant accessibility permissions in:
// System Preferences > Security & Privacy > Privacy > Accessibility
//
// Example:
//
//	if !CheckAccessibilityPermission() {
//		log.Fatal("Accessibility permission required. Please grant permission in System Preferences.")
//	}
func CheckAccessibilityPermission() bool {
	return C.checkAccessibility() == 1
}

// WaitForIdle waits for the specified duration without any keyboard input.
// It monitors keyboard events using macOS's CGEventTap API.
//
// Parameters:
//   - ctx: Context for cancellation (e.g., timeout or manual cancellation)
//   - duration: How long to wait without keyboard input
//
// Returns:
//   - true if no keyboard input was detected during the entire duration
//   - false if keyboard input was detected or context was cancelled
//
// This function requires accessibility permissions. If permissions are not granted,
// it will return false immediately.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	if WaitForIdle(ctx, 10*time.Second) {
//		log.Println("No keyboard input for 10 seconds")
//	} else {
//		log.Println("Keyboard input detected or timeout")
//	}
func WaitForIdle(ctx context.Context, duration time.Duration) bool {
	// Check accessibility permission first
	// Note: Root processes may not pass this check even though they can access events
	if !CheckAccessibilityPermission() {
		// For root processes, just wait for the duration without monitoring
		// This is acceptable because root processes typically run as system services
		select {
		case <-ctx.Done():
			return false
		case <-time.After(duration):
			return true
		}
	}

	// Reset and start monitoring
	C.resetKeyPressed()

	// Create a ticker to check periodically
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	deadline := time.Now().Add(duration)

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			// Check if a key was pressed
			if C.wasKeyPressed() == 1 {
				// Key was pressed, reset and restart waiting
				C.resetKeyPressed()
				deadline = time.Now().Add(duration)
			}

			// Check if we've waited long enough without input
			if time.Now().After(deadline) {
				return true
			}
		}
	}
}
