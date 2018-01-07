#include "driver.h"
#include "json.h"
#include "sandbox.h"
#include "window.h"

@implementation Driver
+ (instancetype)current {
  static Driver *driver = nil;

  @synchronized(self) {
    if (driver == nil) {
      driver = [[Driver alloc] init];
      NSApplication *app = [NSApplication sharedApplication];
      app.delegate = driver;
    }
  }
  return driver;
}

- (instancetype)init {
  self = [super init];

  self.elements = [NSMutableDictionary dictionaryWithCapacity:256];
  self.objc = [[OBJCBridge alloc] init];
  self.golang = [[GoBridge alloc] init];

  // Drivers handlers.
  [self.objc handle:@"/driver/run"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [self run:url payload:payload];
            }];
  [self.objc handle:@"/driver/resources"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [self resources:url payload:payload];
            }];
  [self.objc handle:@"/driver/support"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [self support:url payload:payload];
            }];

  // Window handlers.
  [self.objc handle:@"/window/new"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window newWindow:url payload:payload];
            }];
  [self.objc handle:@"/window/load"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window load:url payload:payload];
            }];
  [self.objc handle:@"/window/render"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window render:url payload:payload];
            }];
  [self.objc handle:@"/window/render/attributes"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window renderAttributes:url payload:payload];
            }];
  [self.objc handle:@"/window/position"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window position:url payload:payload];
            }];
  [self.objc handle:@"/window/move"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window move:url payload:payload];
            }];
  [self.objc handle:@"/window/center"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window center:url payload:payload];
            }];
  [self.objc handle:@"/window/size"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window size:url payload:payload];
            }];
  [self.objc handle:@"/window/resize"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window resize:url payload:payload];
            }];
  [self.objc handle:@"/window/focus"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window focus:url payload:payload];
            }];
  [self.objc handle:@"/window/togglefullscreen"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window toggleFullScreen:url payload:payload];
            }];
  [self.objc handle:@"/window/toggleminimize"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window toggleMinimize:url payload:payload];
            }];
  [self.objc handle:@"/window/close"
            handler:^(NSURLComponents *url, NSString *payload) {
              return [Window close:url payload:payload];
            }];

  self.dock = [[NSMenu alloc] initWithTitle:@""];
  return self;
}

- (bridge_result)run:(NSURLComponents *)url payload:(NSString *)payload {
  [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];
  [NSApp activateIgnoringOtherApps:YES];
  [NSApp run];
  return make_bridge_result(nil, nil);
}

- (bridge_result)resources:(NSURLComponents *)url payload:(NSString *)payload {
  NSBundle *mainBundle = [NSBundle mainBundle];
  NSString *res = [JSONEncoder encodeString:mainBundle.resourcePath];
  return make_bridge_result(res, nil);
}

- (bridge_result)support:(NSURLComponents *)url payload:(NSString *)payload {
  NSBundle *mainBundle = [NSBundle mainBundle];
  NSString *dirname = nil;

  if ([mainBundle isSandboxed]) {
    dirname = [JSONEncoder encodeString:NSHomeDirectory()];
    return make_bridge_result(dirname, nil);
  }

  NSArray *paths = NSSearchPathForDirectoriesInDomains(
      NSApplicationSupportDirectory, NSUserDomainMask, YES);
  NSString *applicationSupportDirectory = [paths firstObject];

  if (mainBundle.bundleIdentifier.length == 0) {
    dirname = [NSString
        stringWithFormat:@"%@/goapp/{appname}", applicationSupportDirectory];
  } else {
    dirname = [NSString stringWithFormat:@"%@/%@", applicationSupportDirectory,
                                         mainBundle.bundleIdentifier];
  }
  dirname = [JSONEncoder encodeString:dirname];
  return make_bridge_result(dirname, nil);
}

- (void)applicationDidFinishLaunching:(NSNotification *)aNotification {

  [self.golang request:@"/driver/run" payload:nil];
}

- (void)applicationDidBecomeActive:(NSNotification *)aNotification {
  [self.golang request:@"/driver/focus" payload:nil];
}

- (void)applicationDidResignActive:(NSNotification *)aNotification {
  [self.golang request:@"/driver/blur" payload:nil];
}

- (BOOL)applicationShouldHandleReopen:(NSApplication *)sender
                    hasVisibleWindows:(BOOL)flag {
  NSString *payload = flag ? @"true" : @"false";
  [self.golang request:@"/driver/reopen" payload:payload];
  return YES;
}

- (void)application:(NSApplication *)sender
          openFiles:(NSArray<NSString *> *)filenames {
  NSString *payload = [JSONEncoder encodeObject:filenames];
  [self.golang request:@"/driver/filesopen" payload:payload];
}

- (void)applicationWillFinishLaunching:(NSNotification *)aNotification {
  NSAppleEventManager *appleEventManager =
      [NSAppleEventManager sharedAppleEventManager];
  [appleEventManager
      setEventHandler:self
          andSelector:@selector(handleGetURLEvent:withReplyEvent:)
        forEventClass:kInternetEventClass
           andEventID:kAEGetURL];
}

- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
  NSString *rawurl =
      [event paramDescriptorForKeyword:keyDirectObject].stringValue;
  NSString *payload = [JSONEncoder encodeString:rawurl];
  [self.golang request:@"/driver/urlopen" payload:payload];
}

- (NSApplicationTerminateReply)applicationShouldTerminate:
    (NSApplication *)sender {
  NSString *res = [self.golang requestWithResult:@"/driver/quit" payload:nil];
  return [JSONDecoder decodeBool:res];
}

- (void)applicationWillTerminate:(NSNotification *)aNotification {
  [self.golang requestWithResult:@"/driver/exit" payload:nil];
}

- (NSMenu *)applicationDockMenu:(NSApplication *)sender {
  return self.dock;
}
@end
