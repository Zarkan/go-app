#include "menu.h"
#include "driver.h"
#include "image.h"
#include "json.h"

@implementation MenuItem
+ (instancetype)create:(NSString *)ID
               compoID:(NSString *)compoID
                inMenu:(NSString *)elemID {
  MenuItem *i =
      [[MenuItem alloc] initWithTitle:@"" action:NULL keyEquivalent:@""];
  i.ID = ID;
  i.compoID = compoID;
  i.elemID = elemID;
  return i;
}

- (void)setAttr:(NSString *)key value:(NSString *)value {
  if ([key isEqual:@"separator"] && self.separator == nil) {
    [self setSeparator];
    return;
  }

  if ([key isEqual:@"label"]) {
    self.title = value != nil ? value : @"";
    return;
  }

  if ([key isEqual:@"disabled"]) {
    self.enabled = NO;
    [self setupOnClick];
    return;
  }

  if ([key isEqual:@"title"]) {
    self.toolTip = value;
    return;
  }

  if ([key isEqual:@"checked"]) {
    self.state = NSControlStateValueOn;
    return;
  }

  if ([key isEqual:@"keys"]) {
    self.keys = value;
    [self setupKeys];
    return;
  }

  if ([key isEqual:@"icon"]) {
    NSString *icon = value;
    icon = icon != nil ? icon : @"";

    if (![self.icon isEqual:icon]) {
      self.icon = icon;
      [self setIconWithPath:icon];
    }
    return;
  }

  if ([key isEqual:@"onclick"]) {
    self.onClick = value;
    [self setupOnClick];
    return;
  }

  if ([key isEqual:@"role"]) {
    self.selector = [[Driver current] selectorFromRole:value];
    [self setupOnClick];
    return;
  }
}

- (void)delAttr:(NSString *)key {
  if ([key isEqual:@"separator"] && self.separator != nil) {
    [self unsetSeparator];
    return;
  }

  if ([key isEqual:@"label"]) {
    self.title = @"";
    return;
  }

  if ([key isEqual:@"disabled"]) {
    self.enabled = YES;
    [self setupOnClick];
    return;
  }

  if ([key isEqual:@"title"]) {
    self.toolTip = nil;
    return;
  }

  if ([key isEqual:@"checked"]) {
    self.state = NSControlStateValueOff;
    return;
  }

  if ([key isEqual:@"keys"]) {
    self.keys = nil;
    [self setupKeys];
    return;
  }

  if ([key isEqual:@"icon"]) {
    NSString *icon = @"";
    self.icon = icon;
    [self setIconWithPath:icon];
    return;
  }

  if ([key isEqual:@"onclick"]) {
    self.onClick = nil;
    [self setupOnClick];
    return;
  }

  if ([key isEqual:@"role"]) {
    self.selector = nil;
    [self setupOnClick];
    return;
  }
}

- (void)setSeparator {
  NSMenuItem *sep = [NSMenuItem separatorItem];
  self.separator = sep;

  MenuContainer *parent = (MenuContainer *)self.menu;
  if (parent == nil) {
    return;
  }

  NSInteger index = [parent indexOfItem:self];
  [parent removeItemAtIndex:index];
  [parent insertItem:sep atIndex:index];
}

- (void)unsetSeparator {
  NSMenuItem *sep = self.separator;
  self.separator = nil;

  MenuContainer *parent = (MenuContainer *)sep.menu;
  if (parent == nil) {
    return;
  }

  NSInteger index = [parent indexOfItem:sep];
  [parent removeItemAtIndex:index];
  [parent insertItem:self atIndex:index];
}

- (void)setIconWithPath:(NSString *)icon {
  if (icon.length == 0) {
    self.image = nil;
    return;
  }

  CGFloat menuBarHeight = [[NSApp mainMenu] menuBarHeight];

  NSImage *img = [[NSImage alloc] initByReferencingFile:icon];
  self.image = [NSImage resizeImage:img
                  toPixelDimensions:NSMakeSize(menuBarHeight, menuBarHeight)];
}

- (void)setupOnClick {
  if (!self.enabled) {
    self.action = nil;
    return;
  }

  if (self.hasSubmenu) {
    self.action = @selector(submenuAction:);
    return;
  }

  if (self.selector != nil) {
    self.action = self.selector;
    return;
  }

  if (self.onClick == nil || self.onClick.length == 0) {
    return;
  }

  self.target = self;
  self.action = @selector(clicked:);
}

- (void)clicked:(id)sender {
  Driver *driver = [Driver current];

  NSDictionary *mapping = @{
    @"CompoID" : self.compoID,
    @"FieldOrMethod" : self.onClick,
    @"JSONValue" : @"{}",
  };

  NSDictionary *in = @{
    @"ID" : self.elemID,
    @"Mapping" : [JSONEncoder encode:mapping],
  };

  [driver.goRPC call:@"menus.OnCallback" withInput:in];
}

- (void)setupKeys {
  if (self.keys == nil || self.keys.length == 0) {
    return;
  }

  self.keyEquivalentModifierMask = 0;
  self.keys = [self.keys lowercaseString];

  NSArray *keys = [self.keys componentsSeparatedByString:@"+"];
  for (NSString *key in keys) {
    if ([key isEqual:@"cmd"] || [key isEqual:@"cmdorctrl"]) {
      self.keyEquivalentModifierMask |= NSEventModifierFlagCommand;
    } else if ([key isEqual:@"ctrl"]) {
      self.keyEquivalentModifierMask |= NSEventModifierFlagControl;
    } else if ([key isEqual:@"alt"]) {
      self.keyEquivalentModifierMask |= NSEventModifierFlagOption;
    } else if ([key isEqual:@"shift"]) {
      self.keyEquivalentModifierMask |= NSEventModifierFlagShift;
    } else if ([key isEqual:@"fn"]) {
      self.keyEquivalentModifierMask |= NSEventModifierFlagFunction;
    } else if ([key isEqual:@""]) {
      self.keyEquivalent = @"+";
    } else {
      self.keyEquivalent = key;
    }
  }
}
@end

@implementation MenuContainer
+ (instancetype)create:(NSString *)ID
               compoID:(NSString *)compoID
                inMenu:(NSString *)elemID {
  MenuContainer *m = [[MenuContainer alloc] initWithTitle:@""];
  m.ID = ID;
  m.compoID = compoID;
  m.elemID = elemID;
  return m;
}

- (void)setAttr:(NSString *)key value:(NSString *)value {
  if ([key isEqual:@"label"]) {
    self.title = value != nil ? value : @"";
  } else if ([key isEqual:@"disabled"]) {
    self.disabled = true;
  }

  [self updateParentItem];
}

- (void)delAttr:(NSString *)key {
  if ([key isEqual:@"label"]) {
    self.title = @"";
  } else if ([key isEqual:@"disabled"]) {
    self.disabled = false;
  }

  [self updateParentItem];
}

- (void)updateParentItem {
  NSMenu *supermenu = self.supermenu;
  if (supermenu == nil) {
    return;
  }

  // Updating parent menuitem title.
  for (NSMenuItem *i in supermenu.itemArray) {
    if (i.submenu == self) {
      i.title = self.title;
      i.enabled = !self.disabled;
      return;
    }
  }
}

- (void)insertChild:(id)child atIndex:(NSInteger)index {
  if ([child isKindOfClass:[MenuContainer class]]) {
    MenuContainer *c = child;
    NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:c.title
                                                  action:NULL
                                           keyEquivalent:@""];

    item.submenu = c;
    item.enabled = !c.disabled;
    [self insertItem:item atIndex:index];
    return;
  }

  MenuItem *item = child;

  if (item.separator != nil) {
    [self insertItem:item.separator atIndex:index];
    return;
  }

  [self insertItem:item atIndex:index];
}

- (void)appendChild:(id)child {
  [self insertChild:child atIndex:self.numberOfItems];
  [self refreshMenuBarOrder];
}

- (void)removeChild:(id)child {
  if ([child isKindOfClass:[MenuContainer class]]) {
    for (NSMenuItem *c in self.itemArray) {
      if (c.submenu == child) {
        [self removeItem:c];
        [self refreshMenuBarOrder];
        return;
      }
    }

    return;
  }

  MenuItem *item = child;

  if (item.separator != nil) {
    [self removeItem:item.separator];
    [self refreshMenuBarOrder];
    return;
  }

  [self removeItem:item];
  [self refreshMenuBarOrder];
}

- (void)replaceChild:(id)old with:(id) new {
  NSInteger index = -1;

  if ([old isKindOfClass:[MenuContainer class]]) {
    NSArray<NSMenuItem *> *children = self.itemArray;

    for (int i = 0; i < children.count; ++i) {
      if (children[i].submenu == old) {
        index = i;
        break;
      }
    }
  } else {
    MenuItem *item = old;

    if (item.separator != nil) {
      index = [self indexOfItem:item.separator];
    } else {
      index = [self indexOfItem:item];
    }
  }

  if (index < 0) {
    return;
  }

  [self removeItemAtIndex:index];
  [self insertChild:new atIndex:index];
}

- (void)refreshMenuBarOrder {
  if (NSApp.mainMenu == self) {
    NSApp.mainMenu = nil;
    NSApp.mainMenu = self;
  }
}
@end

@implementation Menu
- (instancetype)initWithID:(NSString *)ID {
  self = [super init];

  self.ID = ID;
  self.nodes = [[NSMutableDictionary alloc] init];

  return self;
}

+ (void) new:(NSDictionary *)in return:(NSString *)returnID {
  defer(returnID, ^{
    NSString *ID = in[@"ID"];
    Menu *menu = [[Menu alloc] initWithID:ID];

    Driver *driver = [Driver current];
    driver.elements[ID] = menu;
    [driver.macRPC return:returnID withOutput:nil andError:nil];
  });
}

+ (void)load:(NSDictionary *)in return:(NSString *)returnID {
  defer(returnID, ^{
    Driver *driver = [Driver current];
    NSString *ID = in[@"ID"];

    Menu *menu = driver.elements[ID];
    if (menu == nil) {
      [NSException raise:@"ErrNoMenu" format:@"no menu with id %@", ID];
    }

    menu.root = nil;
    [driver.macRPC return:returnID withOutput:nil andError:nil];
  });
}

+ (void)render:(NSDictionary *)in return:(NSString *)returnID {
  defer(returnID, ^{
    Driver *driver = [Driver current];
    NSString *ID = in[@"ID"];
    NSArray *changes = [JSONDecoder decode:in[@"Changes"]];

    Menu *menu = driver.elements[ID];
    if (menu == nil) {
      [NSException raise:@"ErrNoMenu" format:@"no menu with id %@", ID];
    }

    for (NSDictionary *c in changes) {
      NSNumber *action = c[@"Action"];

      switch (action.intValue) {
      case 0:
        [menu setRootNode:c];
        break;

      case 1:
        [menu newNode:c];
        break;

      case 2:
        [menu delNode:c];
        break;

      case 3:
        [menu setAttr:c];
        break;

      case 4:
        [menu delAttr:c];
        break;

      case 6:
        [menu appendChild:c];
        break;

      case 7:
        [menu removeChild:c];
        break;

      case 8:
        [menu replaceChild:c];
        break;

      default:
        [NSException raise:@"ErrChange"
                    format:@"%@ change is not supported", action];
      }
    }

    [driver.macRPC return:returnID withOutput:nil andError:nil];
  });
}

- (void)setRootNode:(NSDictionary *)change {
  NSString *nodeID = change[@"NodeID"];

  MenuCompo *node = self.nodes[nodeID];
  node.isRootCompo = YES;

  id root = [self compoRoot:node];
  if (root == nil) {
    return;
  }

  if (![root isKindOfClass:[MenuContainer class]]) {
    [NSException raise:@"ErrMenu" format:@"menu base is not a menuitem"];
  }

  self.root = root;
}

- (void)newNode:(NSDictionary *)change {
  NSString *nodeID = change[@"NodeID"];
  NSString *compoID = change[@"CompoID"];
  NSString *type = change[@"Type"];
  BOOL isCompo = [change[@"IsCompo"] boolValue];

  if (isCompo) {
    MenuCompo *c = [[MenuCompo alloc] init];
    c.ID = nodeID;
    c.type = type;
    c.isRootCompo = NO;
    self.nodes[nodeID] = c;
    return;
  }

  if ([type isEqual:@"menu"]) {
    self.nodes[nodeID] =
        [MenuContainer create:nodeID compoID:compoID inMenu:self.ID];
    return;
  }

  if ([type isEqual:@"menuitem"]) {
    self.nodes[nodeID] =
        [MenuItem create:nodeID compoID:compoID inMenu:self.ID];
    return;
  }

  [NSException raise:@"ErrMenu" format:@"menu does not support %@ tag", type];
}

- (void)delNode:(NSDictionary *)change {
  [self.nodes removeObjectForKey:change[@"NodeID"]];
}

- (void)setAttr:(NSDictionary *)change {
  id node = self.nodes[change[@"NodeID"]];
  if (node == nil) {
    return;
  }

  NSString *key = change[@"Key"];
  NSString *value = change[@"Value"];

  if ([node isKindOfClass:[MenuContainer class]]) {
    MenuContainer *m = node;
    [m setAttr:key value:value];
    return;
  }

  if ([node isKindOfClass:[MenuItem class]]) {
    MenuItem *mi = node;
    [mi setAttr:key value:value];
    return;
  }

  [NSException raise:@"ErrMenu" format:@"unknown menu element"];
}

- (void)delAttr:(NSDictionary *)change {
  id node = self.nodes[change[@"NodeID"]];
  if (node == nil) {
    return;
  }

  NSString *key = change[@"Key"];

  if ([node isKindOfClass:[MenuContainer class]]) {
    MenuContainer *m = node;
    [m delAttr:key];
    return;
  }

  if ([node isKindOfClass:[MenuItem class]]) {
    MenuItem *mi = node;
    [mi delAttr:key];
    return;
  }

  [NSException raise:@"ErrMenu" format:@"unknown menu element"];
}

- (void)appendChild:(NSDictionary *)change {
  NSString *nodeID = change[@"NodeID"];
  NSString *childID = change[@"ChildID"];

  id node = self.nodes[nodeID];
  if (node == nil) {
    return;
  }

  if ([node isKindOfClass:[MenuCompo class]]) {
    MenuCompo *compo = node;
    compo.rootID = childID;
    return;
  }

  id child = self.nodes[childID];
  child = [self compoRoot:child];
  if (child == nil) {
    return;
  }

  MenuContainer *m = node;
  [m appendChild:child];
}

- (void)removeChild:(NSDictionary *)change {
  NSString *nodeID = change[@"NodeID"];
  NSString *childID = change[@"ChildID"];

  MenuContainer *node = self.nodes[nodeID];
  if (node == nil) {
    return;
  }

  id child = self.nodes[childID];
  child = [self compoRoot:child];
  if (child == nil) {
    return;
  }

  [node removeChild:child];
}

- (void)replaceChild:(NSDictionary *)change {
  NSString *nodeID = change[@"NodeID"];
  NSString *childID = change[@"ChildID"];
  NSString *newChildID = change[@"NewChildID"];

  id node = self.nodes[nodeID];
  if (node == nil) {
    return;
  }

  id child = self.nodes[childID];
  child = [self compoRoot:child];
  if (child == nil) {
    return;
  }

  id newChild = self.nodes[newChildID];
  newChild = [self compoRoot:newChild];
  if (newChild == nil) {
    return;
  }

  if ([node isKindOfClass:[MenuCompo class]]) {
    MenuCompo *compo = node;
    compo.rootID = newChildID;

    if (compo.isRootCompo) {
      [self setRootNode:@{ @"NodeID" : compo.ID }];
    }

    return;
  }

  MenuContainer *m = node;
  [m replaceChild:child with:newChild];
}

- (id)compoRoot:(id)node {
  if (node == nil || ![node isKindOfClass:[MenuCompo class]]) {
    return node;
  }

  MenuCompo *c = node;
  return [self compoRoot:self.nodes[c.rootID]];
}

- (void)menuDidClose:(NSMenu *)menu {
  NSDictionary *in = @{
    @"ID" : self.ID,
  };

  Driver *driver = [Driver current];
  [driver.goRPC call:@"menus.OnClose" withInput:in];
}

+ (void) delete:(NSDictionary *)in return:(NSString *)returnID {
  defer(returnID, ^{
    NSString *ID = in[@"ID"];

    Driver *driver = [Driver current];
    [driver.elements removeObjectForKey:ID];
    [driver.macRPC return:returnID withOutput:nil andError:nil];
  });
}
@end

@implementation MenuCompo
@end