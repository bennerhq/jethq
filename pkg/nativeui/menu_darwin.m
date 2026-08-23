#import <Cocoa/Cocoa.h>

extern void jetkvmNativeMenuAction(char *action);

@interface JetKVMMenuTarget : NSObject
@end

@implementation JetKVMMenuTarget
- (void)performAction:(NSMenuItem *)sender {
  NSString *actionName = (NSString *)sender.representedObject;
  const char *action = actionName.UTF8String;
  if (action != NULL) {
    jetkvmNativeMenuAction((char *)action);
  }
}
@end

static JetKVMMenuTarget *jetkvmMenuTarget;
static BOOL jetkvmMenuInstalled = NO;

void jetkvmSetApplicationIcon(const unsigned char *data, size_t length) {
  NSData *iconData = [[NSData alloc] initWithBytes:data length:length];
  dispatch_async(dispatch_get_main_queue(), ^{
    NSApplication *application = NSApp ?: [NSApplication sharedApplication];
    NSImage *icon = [[NSImage alloc] initWithData:iconData];
    if (icon != nil) {
      [application setApplicationIconImage:icon];
    }
  });
}

static void addItem(NSMenu *menu, NSString *title, NSString *action, NSString *key) {
  NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:@selector(performAction:) keyEquivalent:key];
  item.target = jetkvmMenuTarget;
  item.representedObject = action;
  [menu addItem:item];
}

static void installSettingsMenuWhenReady(void) {
    if (jetkvmMenuInstalled) return;

    NSApplication *application = NSApp ?: [NSApplication sharedApplication];
    NSMenu *mainMenu = application.mainMenu;
    // Ebiten/GLFW creates its menu between Cocoa's will-finish-launching and
    // did-finish-launching callbacks. This bridge is called before RunGame,
    // so wait until that lifecycle work has supplied the native menu bar.
    if (mainMenu == nil) {
      dispatch_after(dispatch_time(DISPATCH_TIME_NOW, (int64_t)(100 * NSEC_PER_MSEC)), dispatch_get_main_queue(), ^{
        installSettingsMenuWhenReady();
      });
      return;
    }
    if ([mainMenu itemWithTitle:@"Tools"] != nil) {
      jetkvmMenuInstalled = YES;
      return;
    }

    jetkvmMenuTarget = [JetKVMMenuTarget new];
    NSMenu *settings = [[NSMenu alloc] initWithTitle:@"Tools"];
    addItem(settings, @"Reconnect", @"reconnect", @"r");
    [settings addItem:[NSMenuItem separatorItem]];
    addItem(settings, @"Paste Text", @"paste", @"p");
    addItem(settings, @"Virtual Media", @"media", @"m");
    addItem(settings, @"Serial Console", @"serial_console", @"t");
    [settings addItem:[NSMenuItem separatorItem]];
    addItem(settings, @"Connection Stats", @"stats", @"i");
    addItem(settings, @"Toggle Full Screen", @"fullscreen", @"f");
    [settings addItem:[NSMenuItem separatorItem]];
    addItem(settings, @"Settings…", @"settings", @",");

    NSMenuItem *root = [[NSMenuItem alloc] initWithTitle:@"Tools" action:nil keyEquivalent:@""];
    root.submenu = settings;
    [mainMenu addItem:root];
    jetkvmMenuInstalled = YES;
}

void jetkvmInstallSettingsMenu(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    installSettingsMenuWhenReady();
  });
}
