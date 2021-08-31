
#import "runloop_nsapp.h"
#include "_cgo_export.h"

@interface ContentView: NSView
@end

ContentView *contentView;

@interface AppDelegate: NSObject <NSApplicationDelegate>
@end

@implementation AppDelegate {
  NSWindow *mainWindow;
}

- (void)applicationWillFinishLaunching:(NSNotification *)note {
    NSLog(@"Will finish launching");
}

- (void) applicationDidFinishLaunching:(NSNotification *)note {
    NSLog(@"Did finish launching");
    [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

    // Menubar
    id menubar = [[NSMenu new] autorelease];
    id appMenuItem = [[NSMenuItem new] autorelease];
    id appMenu = [[NSMenu new] autorelease];
    id appName = [NSProcessInfo processInfo].processName;
    id quitTitle = [@"Quit " stringByAppendingString:appName];
    id quitMenuItem = [[[NSMenuItem alloc] initWithTitle:quitTitle
        action:@selector(terminate:) keyEquivalent:@"q"]
            autorelease];

    [NSApp setMainMenu:menubar];
    [menubar addItem:appMenuItem];
    [appMenuItem setSubmenu:appMenu];
    [appMenu addItem:quitMenuItem];

    // Create the window and content view
    mainWindow = [[NSWindow alloc]
        initWithContentRect:contentView.frame
        styleMask: NSWindowStyleMaskClosable | NSWindowStyleMaskTitled
        backing:NSBackingStoreBuffered
        defer:NO];

    [mainWindow cascadeTopLeftFromPoint:NSMakePoint(20,20)];
    [mainWindow setTitle:appName];

    mainWindow.contentView = contentView;
    
    [mainWindow makeKeyAndOrderFront:nil];
    [NSApp activateIgnoringOtherApps:YES];
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
  return YES;
}

@end

@implementation ContentView {
}

- (id) initWithFrame:(CGRect)frame {
  if (self = [super initWithFrame:frame]) {
    self.wantsLayer = YES;
    self.layerContentsRedrawPolicy = NSViewLayerContentsRedrawNever;
    self.layer.magnificationFilter = kCAFilterNearest;
  }
  return self;
}

- (void) setImage:(id)image {
  dispatch_async(dispatch_get_main_queue(), ^{
    self.layer.contents = image;
  });
}

- (void) forwardMouseEvent: (NSEvent *)event {
  NSPoint pos = [self convertPoint: event.locationInWindow fromView:nil];
  BOOL mouseDown = NSEvent.pressedMouseButtons & 1;
  int x = pos.x;
  int y = self.frame.size.height - pos.y;
  // Call the Go function that handles mouse events.
  receiveMouseEvent(x, y, mouseDown);
}

- (void) mouseDown: (NSEvent *)event {
  [self forwardMouseEvent:event];
}

- (void) mouseUp: (NSEvent *)event {
  [self forwardMouseEvent:event];
}

- (void) mouseDragged: (NSEvent *)event {
  [self forwardMouseEvent:event];
}

@end

void InitApp(int w, int h) {
  @autoreleasepool {
    // Create the content view with correct width and height
    contentView = [[ContentView alloc] initWithFrame: NSMakeRect(0, 0, w, h)];
    // Constant size for the content view
    contentView.autoresizingMask = NSViewNotSizable;
  }
}

void RunApp() {
  @autoreleasepool {
    // Create and assign the app delegate, and run it.
    [NSApplication sharedApplication];
    NSApp.delegate = [AppDelegate new];
    [NSApp run];
  }
}

void StopApp() {
  @autoreleasepool {
  	[NSApp terminate:nil];
  }
}

void DrawRGBA(void* bytes, int length, int w, int h) {
  CGDataProviderRef provider = CGDataProviderCreateWithData(NULL, bytes, length, nil);
  
  CGColorSpaceRef colorspace = CGColorSpaceCreateDeviceRGB();
  CGImageRef image = CGImageCreate(
      w,
      h,
      8,
      32,
      4*w,
      colorspace,
      kCGBitmapByteOrderDefault,
      provider,
      NULL,
      false,
      kCGRenderingIntentDefault
  );

  [contentView setImage: (__bridge id)image];

  CGImageRelease(image);
  CGColorSpaceRelease(colorspace);
  CGDataProviderRelease(provider);
}
