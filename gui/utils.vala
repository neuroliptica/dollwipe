// utils.vala: definitions to create and manipulate widgets.

namespace Utils {

using Consts;

Gtk.Window create_main_window() {
    var window = new Gtk.Window();
    
    window.set_title(MainWindow.TITLE);
    window.set_default_size(MainWindow.WIDTH, MainWindow.HEIGHT);

    window.destroy.connect((a) => {
        Gtk.main_quit();
    });

    return window;
}

Gtk.Box create_hbox(bool homogeneous, int spacing = 0) {
    var hbox = new Gtk.Box(Gtk.Orientation.HORIZONTAL, spacing);
    hbox.set_homogeneous(homogeneous);

    return hbox;
}

Gtk.Box create_vbox(bool homogeneous, int spacing = 0) {
    var vbox = new Gtk.Box(Gtk.Orientation.VERTICAL, spacing);
    vbox.set_homogeneous(homogeneous);

    return vbox;
}

Gtk.Paned create_vpaned() {
    return new Gtk.Paned(Gtk.Orientation.VERTICAL);
}

Gtk.Paned create_hpaned() {
    return new Gtk.Paned(Gtk.Orientation.HORIZONTAL);
}

Gtk.Frame create_frame_label_center(string? label) {
    var frame = new Gtk.Frame(label);
    frame.set_label_align(0.5f, 0);

    return frame;
}

Gtk.RadioButton[] create_radio_buttons(string[] labels) {
    var buttons = new Gtk.RadioButton[labels.length];
    for (int i = 0; i < labels.length; ++i) {
        Gtk.RadioButton? group = (i == 0) ? null : buttons[0];
        var button = new Gtk.RadioButton.with_label_from_widget(group, labels[i]);
        buttons[i] = button;
    }
    
    return buttons;
}

//Gtk.Box create_vbox_with_parent(bool    homogeneous,
//                                int     spacing,
//                                Gtk.Box parent,
//                                int     width,
//                                int     height)
//{
//    Gtk.Box vbox = create_vbox(homogeneous, spacing);
//    vbox.set_size_request(width, height);
//    parent.add(vbox);
//
//    return vbox;
//}

Gtk.Box create_hbox_with_parent(bool    homogeneous,
                                int     spacing,
                                Gtk.Box parent,
                                int     width,
                                int     height)
{
    Gtk.Box hbox = create_hbox(homogeneous, spacing);
    hbox.set_size_request(width, height);
    parent.add(hbox);

    return hbox;
}

string remove_trailing_spaces(string src) {
    string result = "";
    for (int i = 0; i < src.length; ++i) {
        if (src[i] == ' ' || src[i] == '\n')
            break;
        result += @"$(src[i])";
    }
    return result;
}

bool is_digit_string(string src) {
    for (int i = 0; i < src.length; ++i)
        if (!src[i].isdigit())
            return false;
    return true;
}

} // Utils
