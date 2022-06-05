// base.vala: main widget classes, from which kuklogui widgets inherits.

namespace GUI {

using Utils;
using Consts;

// Base frame with framebox class.
public class Frame {
    public Gtk.Frame frame { get; set; }
    public Gtk.Box framebox { get; set; }

    public void add_child(Gtk.Widget widget) {
        this.framebox.add(widget);
    }

    // This can probably be set in a base constructor.
    public void add_parent(Gtk.Container widget) {
        widget.add(this.frame);
    }

    public Frame(string? label, bool homogeneous, int spacing, Gtk.Orientation o) {
        this.frame = create_frame_label_center(label);
        this.framebox = new Gtk.Box(o, spacing);
        this.framebox.set_homogeneous(homogeneous);

        this.frame.add(this.framebox);
    }
}

// Base class for frames with just radio buttons.
public class RadioFrame : Frame {
    protected string[] modes { get; set; }
    protected Gtk.RadioButton[] buttons { get; set; }

    protected void create_buttons_from_modes() {
        buttons = create_radio_buttons(this.modes);
        for (int i = 0; i < this.modes.length; ++i)
            this.add_child(buttons[i]);
    }

    public RadioFrame(string? label, bool homogeneous, int spacing, Gtk.Orientation o) {
        base(label, homogeneous, spacing, o);
    }

    public int get_active() {
        for (int i = 0; i < buttons.length; ++i)
            if (buttons[i].active)
                return i;
        return -1; // Empty
    }
}

// Base class for frames with just single text entry.
public class EntryFrame : Frame {
    protected Gtk.Entry entry;

    public bool editable {
        get { return entry.editable; }
        set { entry.editable = value; }
    }
    public string content {
        get { return entry.text; }
        set { entry.text = value; }
    }

    public EntryFrame(string? label, bool homogeneous, int spacing, Gtk.Orientation o) {
        base(label, homogeneous, spacing, o);

        entry = new Gtk.Entry();
        this.add_child(entry);
    }

    protected void set_margin(int start, int end) {
        entry.set_margin_end(end);
        entry.set_margin_start(start);
    }
}

// Frame with just FileChooser widget.
public class FileChooserButtonFrame : Frame {
    protected Gtk.FileChooserButton chooser;

    public FileChooserButtonFrame(string?               label,
                                  bool                  homogeneous,
                                  int                   spacing,
                                  Gtk.Orientation       o,
                                  string                title,
                                  Gtk.FileChooserAction act) 
    {
        base(label, homogeneous, spacing, o);

        chooser = new Gtk.FileChooserButton(title, act);
        // frame.set_label_align(0.1f, 0);
        this.add_child(chooser);
    }

    public string get_filename() {
        string? res = chooser.get_filename();
        return (res == null) ? "" : (string)res;
    }

    protected void set_margin(int start, int end) {
        chooser.set_margin_end(end);
        chooser.set_margin_start(start);
    }
}

// Frame with just SpinButton widget.
public class SpinButtonFrame : Frame {
    protected Gtk.SpinButton button;

    public SpinButtonFrame(string?         label,
                           bool            homogeneous,
                           int             spacing,
                           Gtk.Orientation o,
                           double          min,
                           double          max,
                           double          step)
    {
        base(label, homogeneous, spacing, o);

        button = new Gtk.SpinButton.with_range(min, max, step);
        this.add_child(button);
    }

    public int get_int_value() {
        return (int)button.get_value();
    }

    protected void set_margin(int start, int end) {
        button.set_margin_end(end);
        button.set_margin_start(start);
    }
}

} // GUI
