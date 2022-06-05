// fails.vala: fail windows and general fail controller for kuklogui.

namespace Fail {

using Consts;
using Utils;

// Base window for errors notifications.
class ErrorWindow : Gtk.Window { 
    public    Gtk.Box    vbox;
    protected Gtk.Button ok;
    protected Gtk.Label  content;

    public ErrorWindow(string title, int width, int height) {
        this.set_title(title);
        this.set_default_size(width, height);

        vbox = create_vbox(true, 0);
        this.add(vbox);

        content = new Gtk.Label(null);
        vbox.add(content);

        content.set_margin_start(20);
        content.set_margin_end(20);

        ok = new Gtk.Button.with_label("Пынял");
        vbox.add(ok);

        ok.clicked.connect((a) => {
            this.close();
        });

        ok.set_valign(Gtk.Align.END);
        ok.set_halign(Gtk.Align.END);

        ok.set_margin_end(20);
        ok.set_margin_bottom(20);
    }

    public void set_content(string _content) {
        content.set_text(_content);
    }
}

// If some assertion has failed, then we call one of FailController's methods.
// Every time we call any method -- we should block parent before any other action.
class FailController {
    private Gtk.Window parent;

    public FailController(Gtk.Window _parent) {
        parent = _parent;
    }

    private void block_parent() {
        parent.set_sensitive(false);
    }

    private void unblock_parent() {
        parent.set_sensitive(true);
    }

    private void error_with_content(string name, string content) {
        block_parent();
        var win = new ErrorWindow(name, 300, 80);
        win.set_content(content);
        win.destroy.connect((a) => {
            unblock_parent();
        });
        win.show_all();
    }

    // =========== errors =======
    public void zero_files() {
        string cont = "Для создания тредов необходимо прикрепить хотя бы один файл.";
        error_with_content("Zero files", cont); 
    }

    public void no_thread_id() {
        string cont = "Режим \"один тред\", но id треда не указан.";
        error_with_content("No thread id", cont);
    }

    public void id_wrong_format() {
        string cont = "Id треда указан в неверном формате.";
        error_with_content("Id wrong format", cont);
    }

    public void no_path_media() {
        string cont = "Для прикрепления файлов укажите путь до папки с ними.";
        error_with_content("No path media", cont);
    }

    public void no_path_text() {
        string cont = "Для выбора текста постов из файла укажите путь до файла.";
        error_with_content("No path text", cont);
    }

    public void no_path_proxy() {
        string cont = "Для вайпа с проксями укажите путь до файла с ними.";
        error_with_content("No path proxy", cont);
    }

    public void no_key() {
        string cont = "Чтобы решать капчу укажите ключ антикапчи или пасскод.";
        error_with_content("No key", cont);
    }

    // Block for /rm/, /pr/, /sci/ and /math/
    public void no_wipe_for_this_board() {
        string cont = "Извините, эту доску вайпать нельзя, она защищена магическим полем! Такие дела.";
        error_with_content("Whitelist board.", cont);
    }
}

} // Fail
