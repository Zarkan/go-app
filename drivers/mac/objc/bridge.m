#include "bridge.h"
#include "_cgo_export.h"
#include "driver.h"
#include "json.h"

void macCall(char *rawCall) {
  NSDictionary *call =
      [JSONDecoder decode:[NSString stringWithUTF8String:rawCall]];

  NSString *method = call[@"Method"];
  id in = call[@"In"];
  NSString *returnID = call[@"ReturnID"];

  @try {
    Driver *driver = [Driver current];
    MacRPCHandler handler = driver.macRPC.handlers[method];

    if (handler == nil) {
      [NSException raise:@"not supported" format:@"not supported"];
    }

    handler(in, returnID);
  } @catch (NSException *exception) {
    NSString *err = exception.reason;
    macCallReturn((char *)returnID.UTF8String, nil, (char *)err.UTF8String);
  }
}

void defer(NSString *returnID, dispatch_block_t block) {
  dispatch_async(dispatch_get_main_queue(), ^{
    @try {
      block();
    } @catch (NSException *exception) {
      NSString *err = exception.reason;
      macCallReturn((char *)returnID.UTF8String, nil, (char *)err.UTF8String);
    }
  });
}

@implementation MacRPC
- (instancetype)init {
  self = [super init];
  self.handlers = [NSMutableDictionary dictionaryWithCapacity:64];
  return self;
}

- (void)handle:(NSString *)method withHandler:(MacRPCHandler)handler {
  self.handlers[method] = handler;
}

- (void) return:(NSString *)returnID
     withOutput:(id)out
       andError:(NSString *)err {

  char *creturnID = returnID != nil ? (char *)returnID.UTF8String : nil;
  char *cout = out != nil ? (char *)[JSONEncoder encode:out].UTF8String : nil;
  char *cerr = err != nil ? (char *)err.UTF8String : nil;

  macCallReturn(creturnID, cout, cerr);
}
@end

@implementation GoRPC
- (void)call:(NSString *)method withInput:(id)in {
  NSMutableDictionary *call = [[NSMutableDictionary alloc] init];
  call[@"Method"] = method;
  call[@"In"] = in;

  NSString *callString = [JSONEncoder encode:call];
  goCall((char *)callString.UTF8String);
}
@end