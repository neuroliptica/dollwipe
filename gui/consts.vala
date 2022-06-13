// consts.vala: constant GUI settings.

namespace Consts {

class MainWindow {
    public static string TITLE = "Kuklowipe v1.0.0";
    public static int HEIGHT = 575;
    public static int WIDTH  = 590;
}

class LeftPanedBoxes {
    public static int HEIGHT = 140;
    public static int WIDTH = 290;
}

class EntrySize {
    public static int LEN = 10;
}

class Button {
    public static int MARGIN = 20;
    public static int START_HEIGHT = 45;
}

class Chooser {
    public static int EMARGIN = 10;
    public static int SMARGIN = 10;
}

[CCode(cname = "pixbuf", cheader_filename = "res/header.h")]
public extern Gdk.Pixbuf pixbuf();

} // Consts
