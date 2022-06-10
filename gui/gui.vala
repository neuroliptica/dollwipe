// gui.vala: main kuklogui class and it's widgets.

using Consts;
using Utils;

class WipeMode : GUI.RadioFrame {
    private string[] names = {
        "Один тред", // 0
        "Шрапнель",  // 1
        "Создание"   // 2
    };

    public WipeMode() {
        base("Режим", false, 1, Gtk.Orientation.VERTICAL);
        modes = names;
        create_buttons_from_modes();
    }
}

class AntiCaptcha : GUI.RadioFrame {
    private string[] names = {
        "RuCaptcha",   // 0
        "XCaptcha",    // 1
        "AntiCaptcha", // 2
        "Пасскод"      // 3
       // "Вручную"    // 4
    };

    public AntiCaptcha() {
        base("Антикапча", false, 1, Gtk.Orientation.VERTICAL);
        modes = names;
        create_buttons_from_modes();

        buttons[1].set_sensitive(false); // XCaptcha
        buttons[2].set_sensitive(false); // AntiCaptcha
        buttons[3].set_sensitive(false); // Passcode
    }
}

class ProxyMode : GUI.RadioFrame {
    private string[] names = {
        "Без проксей", // 0
        "Из файла"     // 1
    };

    public ProxyMode() {
        base("Прокси", false, 1, Gtk.Orientation.VERTICAL);
        modes = names;
        create_buttons_from_modes();
    }
}

class TextMode : GUI.RadioFrame {
    private string[] names = {
        "Из файла",   // 0
        "Без текста", // 1
        "Шизобред",   // 2
        "Из постов"   // 3
    };

    public TextMode() {
        base("Текст", false, 1, Gtk.Orientation.VERTICAL);
        modes = names;
        create_buttons_from_modes();

        // set Шизобред to false cuz not implemented yet.
        buttons[2].set_sensitive(false);
    }
}

class Board : GUI.EntryFrame {
    public Board() {
        base("Доска", true, 1, Gtk.Orientation.VERTICAL);

        entry.set_alignment(0.5f);
        entry.set_width_chars(EntrySize.LEN);
        entry.set_placeholder_text("b по дефолту");
        
        set_margin(5, 5);
    }
}

class Thread : GUI.EntryFrame {
    public Thread() {
        base("Тред (id)", true, 1, Gtk.Orientation.VERTICAL);

        entry.set_alignment(0.5f);
        entry.set_width_chars(EntrySize.LEN);
        entry.set_placeholder_text("Если один тред");
    
        set_margin(5, 5);
    }
}

class Files : GUI.SpinButtonFrame {
    public Files() {
        base("Файлы", false, 1, Gtk.Orientation.VERTICAL, 0, 4, 1);

        set_margin(5, 5);
    }
}

class Sage : GUI.Frame {
    private Gtk.CheckButton sage;

    public Sage() {
        base("Сага", false, 0, Gtk.Orientation.VERTICAL);
        
        sage = new Gtk.CheckButton.with_mnemonic("sage");
        this.add_child(sage);
    }

    public bool need_sage() {
        return sage.get_active();
    }
}

class StartButton : GUI.Frame {
    public Gtk.Button button;

    public StartButton() {
        base(null, true, 0, Gtk.Orientation.VERTICAL);

        button = new Gtk.Button.with_label("~Nipa!");
        button.set_size_request(-1, Button.START_HEIGHT);
        button.set_valign(Gtk.Align.CENTER);
        button.set_margin_end(Button.MARGIN);
        button.set_margin_start(Button.MARGIN);

        this.add_child(button);
    }
}

class FilesChooser : GUI.FileChooserButtonFrame {
    public FilesChooser() {
        base("Директория с медиа", true, 0, Gtk.Orientation.VERTICAL, "Файлы", Gtk.FileChooserAction.SELECT_FOLDER);

        set_margin(Chooser.EMARGIN, Chooser.SMARGIN);
    }
}

class TextChooser : GUI.FileChooserButtonFrame {
    public TextChooser() {
        base("Тексты постов", true, 0, Gtk.Orientation.VERTICAL, "Тексты", Gtk.FileChooserAction.OPEN);

        set_margin(10, 10);
    }
}

class ProxyChooser : GUI.FileChooserButtonFrame {
    public ProxyChooser() {
        base("Проксичи", true, 0, Gtk.Orientation.VERTICAL, "Прокси", Gtk.FileChooserAction.OPEN);

        set_margin(10, 10);
    }
}

class Threads : GUI.SpinButtonFrame {
    public Threads() {
        base("Потоки", true, 0, Gtk.Orientation.VERTICAL, 1, 9999, 1);
    }
}

class Iters : GUI.SpinButtonFrame {
    public Iters() {
        base("Проходов", true, 0, Gtk.Orientation.VERTICAL, 1, 9999, 1);
    }
}

class Timeout : GUI.SpinButtonFrame {
    public Timeout() {
        base("Перерыв", true, 0, Gtk.Orientation.VERTICAL, 0, 9999, 1);
    }
}

class Domain : GUI.RadioFrame {
    private string[] names = {
        "2ch.life",
        "2ch.hk"
    };

    public Domain() {
        base("Зеркало", false, 5, Gtk.Orientation.HORIZONTAL);
        
        modes = names;
        create_buttons_from_modes();
    }
}

class AntiCaptchaKey : GUI.EntryFrame {
    public AntiCaptchaKey() {
        base("Ключ / Пасскод", true, 1, Gtk.Orientation.VERTICAL);

        entry.set_alignment(0.5f);
        entry.set_width_chars(EntrySize.LEN);
        entry.set_placeholder_text("API ключ антикапчи или пасскод");
        entry.set_visibility(false);
        entry.set_valign(Gtk.Align.START);

        set_margin(10, 10);
    }
}

class Extra : GUI.Frame {
    private Gtk.CheckButton verbose;
    private Gtk.CheckButton color;

    public Extra() {
        base("Дополнительно", false, 0, Gtk.Orientation.HORIZONTAL);
        
        verbose = new Gtk.CheckButton.with_mnemonic("Доп. логи");
        this.add_child(verbose);

        color = new Gtk.CheckButton.with_mnemonic("Красить картинки");
        this.add_child(color);
    }

    public bool need_verbose() {
        return verbose.get_active();
    }

    public bool need_color() {
        return color.get_active();
    }
}

class Init {
    public int wipe_mode;
    public int text_mode;
    public int captcha_mode;

    public bool need_proxy;
    public bool need_sage;
    public int  files;

    public string files_path;
    public string text_path;
    public string proxy_path;

    public int threads;
    public int iters;
    public int timeout;

    public string board;
    public string thread;

    public string key;
    public string domain;

    public bool verbose;
    public bool need_color;

    private string[] banned = {
        "rm",
        "pr",
        "math",
        "sci"
    };

    private Fail.FailController controller;

    public Init(Gtk.Window parent) {
        controller = new Fail.FailController(parent);
    }

    public bool validate() {
        if (board in banned) {
            controller.no_wipe_for_this_board();
            return false;
        }
        // 0 - один тред
        if (wipe_mode == 0 && !is_digit_string(thread)) {
            controller.id_wrong_format();
            return false;
        }
        // 0 - один тред
        if (wipe_mode == 0 && thread == "") {
            controller.no_thread_id();
            return false;
        }
        // 0 - из файла
        if (text_mode == 0 && text_path == "") {
            controller.no_path_text();
            return false;
        }
        // 2 - создание тредов
        if (wipe_mode == 2 && files == 0) {
            controller.zero_files();
            return false;
        }
        if (files != 0 && files_path == "") {
            controller.no_path_media();
            return false;
        }
        if (need_proxy && proxy_path == "") {
            controller.no_path_proxy();
            return false;
        }
        if (key == "") {
            controller.no_key();
            return false;
        }
        return true;
    }

    public void init_wipe() {
        string command = @"./dollwipe -mode $wipe_mode -text $text_mode -board $board -captcha $captcha_mode -files $files -t $threads -i $iters -key $key -domain $domain";
        if (need_sage)
            command += " -sage";
        if (need_proxy)
            command += " -proxy";
        if (wipe_mode == 0)
            command += @" -thread $thread";
        if (files_path != "")
            command += @" -file-path $files_path";
        if (text_path != "")
            command += @" -caption-path $text_path";
        if (proxy_path != "")
            command += @" -proxy-path $proxy_path";
        if (timeout > 0)
            command += @" -timeout $timeout";
        if (verbose)
            command += " -v";
        if (need_color)
            command += " -color";

        stdout.printf("%s\n", command);
        Posix.system(command);
    }
}

class KukloGUI : GLib.Object {
    public static int main(string[] argv) {
        Gtk.init(ref argv);

        Gtk.Window main_window = create_main_window();

        Gtk.Paned main_vpaned = create_vpaned();
        main_window.add(main_vpaned);

        Gtk.Image header = new Gtk.Image.from_file("./res/header.png");
        main_vpaned.pack1(header, false, false);

        Gtk.Paned settings_hpaned = create_hpaned(); 
        main_vpaned.pack2(settings_hpaned, true, false);

        Gtk.Box left_vbox = create_vbox(false, 0);
        Gtk.Box right_vbox = create_vbox(false, 0);

        settings_hpaned.pack1(left_vbox, false, false);
        settings_hpaned.pack2(right_vbox, true, false);

        // =================== LEFT ====================
        // mode/proxymode
        Gtk.Box modes_hbox = create_hbox_with_parent(true, 0, left_vbox, LeftPanedBoxes.WIDTH, LeftPanedBoxes.HEIGHT-30);
        var wipe_mode_frame = new WipeMode();
        wipe_mode_frame.add_parent(modes_hbox);

        var proxy_frame = new ProxyMode();
        proxy_frame.add_parent(modes_hbox);

        // text/anticaptcha
        Gtk.Box settings_hbox = create_hbox_with_parent(true, 0, left_vbox, LeftPanedBoxes.WIDTH, LeftPanedBoxes.HEIGHT);
        var text_frame = new TextMode();
        text_frame.add_parent(settings_hbox);

        var captcha_frame = new AntiCaptcha();
        captcha_frame.add_parent(settings_hbox);

        // thread/board
        Gtk.Box meta_box = create_hbox_with_parent(true, 0, left_vbox, LeftPanedBoxes.WIDTH, 20);
        var board_frame = new Board();
        board_frame.add_parent(meta_box);

        var thread_frame = new Thread();
        thread_frame.add_parent(meta_box);

        // files/sage
        Gtk.Box files_box = create_hbox_with_parent(true, 0, left_vbox, LeftPanedBoxes.WIDTH, 20);
        var files_frame = new Files();
        files_frame.add_parent(files_box);

        var sage_frame = new Sage();
        sage_frame.add_parent(files_box);

        // start button
        Gtk.Box start_box = create_hbox_with_parent(true, 0, left_vbox, LeftPanedBoxes.WIDTH, 80);
        var start_frame = new StartButton();
        start_frame.add_parent(start_box);
        // =================== /LEFT ===================

        // =================== RIGHT ==================
        // domain
        Gtk.Box domain_box = create_hbox_with_parent(true, 0, right_vbox, -1, -1);
        var domains = new Domain();
        domains.add_parent(domain_box);

        // path for media
        Gtk.Box files_chooser_box = create_hbox_with_parent(true, 0, right_vbox, LeftPanedBoxes.WIDTH, 20);
        var files_chooser_frame = new FilesChooser();
        files_chooser_frame.add_parent(files_chooser_box);

        // text/proxy pathes
        Gtk.Box text_chooser_box = create_hbox_with_parent(true, 0, right_vbox, LeftPanedBoxes.WIDTH, 20);
        var text_chooser_frame = new TextChooser();
        text_chooser_frame.add_parent(text_chooser_box);

        var proxy_chooser_frame = new ProxyChooser();
        proxy_chooser_frame.add_parent(text_chooser_box);

        // wipe settings
        Gtk.Box wipe_box = create_hbox_with_parent(true, 0, right_vbox, LeftPanedBoxes.WIDTH, 20);
        var threads = new Threads();
        var iters   = new Iters();
        var timeout = new Timeout();

        threads.add_parent(wipe_box);
        iters.add_parent(wipe_box);
        timeout.add_parent(wipe_box);

        // api-key/passcode
        Gtk.Box key_box = create_hbox_with_parent(true, 0, right_vbox, LeftPanedBoxes.WIDTH, -1);
        var key         = new AntiCaptchaKey();
        key.add_parent(key_box);

        // extra options
        Gtk.Box extra_box = create_hbox_with_parent(true, 0, right_vbox, LeftPanedBoxes.WIDTH, -1);
        var extra         = new Extra();
        extra.add_parent(extra_box);
        // =================== /RIGHT =================

        main_window.show_all();
            
        Init init = new Init(main_window);

        start_frame.button.clicked.connect((a) => {
            init.wipe_mode    = wipe_mode_frame.get_active();
            init.text_mode    = text_frame.get_active();
            init.captcha_mode = captcha_frame.get_active();

            init.need_proxy = (proxy_frame.get_active() == 1);
            init.need_sage  = sage_frame.need_sage();

            init.files = files_frame.get_int_value();

            init.board = remove_trailing_spaces(board_frame.content);
            if (init.board == "") {
                init.board = "b";
            }
            init.thread = remove_trailing_spaces(thread_frame.content);

            init.files_path = files_chooser_frame.get_filename();
            init.text_path  = text_chooser_frame.get_filename();
            init.proxy_path = proxy_chooser_frame.get_filename();

            init.key    = remove_trailing_spaces(key.content);
            init.domain = (domains.get_active() == 0) ? "life" : "hk";

            init.threads = threads.get_int_value();
            init.iters   = iters.get_int_value();
            init.timeout = timeout.get_int_value();

            init.verbose    = extra.need_verbose();
            init.need_color = extra.need_color();

            if (!init.validate())
                return;

            init.init_wipe();
        });
        
        Gtk.main();
        return 0;
    }
}
