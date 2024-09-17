#include <gtk/gtk.h>
#include <JavaScriptCore/JavaScript.h>
#include <webkit2/webkit2.h>
#include <libayatana-appindicator/app-indicator.h>
#include <string.h>

extern void go_menu_callback(GtkMenuItem *,int);

extern void go_webview_callback(WebKitUserContentManager *manager, WebKitJavascriptResult *r, int arg);

extern void go_event_callback(GtkWindow *window, GdkEvent *event, int arg);


static void _g_signal_connect(GtkWidget *item, char *action, void *callback, int user) {
  g_signal_connect(item, action, G_CALLBACK(callback), (void *)&user);
}