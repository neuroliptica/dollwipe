// env.go: main global environment.
// Processing user input and building main Env struct.

package env

import (
	"dollwipe/logger"
	"dollwipe/network"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
)

// WipeMode: -mode flag consts.
const (
	SINGLE = iota
	SHRAPNEL
	CREATING
)

// AntiCaptcha: -captcha flag consts.
const (
	RUCAPTCHA = iota
	OCR
	XCAPTCHA
	ANTICATPCHA
	PASSCODE
	MANUAL
)

// TextMode: -text flag consts.
const (
	FROM_FILE = iota
	NO_CAPS
	SCHIZO
	FROM_POSTS
	DEFAULT
)

// Mirrors: -domain flag consts.
var domains = map[string]bool{
	"hk": true,
	// "life": true,
}

// Banned boards, block for -board flag.
var banned = map[string]bool{
	"rm":   true,
	"math": true,
	"pr":   true,
	"sci":  true,
}

var (
	// General wipe mode settings.
	wipeMode       = flag.Uint64("mode", SHRAPNEL, "режим вайпа:\n\t0 - один тред\n\t1 - шрапнель\n\t2 - создание")
	textMode       = flag.Uint64("text", FROM_FILE, "тексты постов:\n\t0 - брать из файла\n\t1 - без текста\n\t2 - шизобред\n\t3 - из постов\n\t4 - дефолтные")
	antiCaptcha    = flag.Uint64("captcha", OCR, "антикапча:\n\t0 - RuCaptcha\n\t1 - OCR")
	antiCaptchaKey = flag.String("key", "", "ключ API антикапчи, либо пасскод.")

	// Post settings.
	board    = flag.String("board", "b", "доска.")
	thread   = flag.Uint64("thread", 0, "ID треда, если вайпаем один тред.")
	files    = flag.Uint64("files", 0, "кол-во прикрепляемых файлов.")
	useProxy = flag.Bool("proxy", false, "вайпать с проксями.")
	useSage  = flag.Bool("sage", false, "клеить сажу.")
	colorize = flag.Bool("masks", false, "цветовые маски для картинок.")

	// Path settings.
	filesPath = flag.String("file-path", "./res/files/", "директория с файлами.")
	capsPath  = flag.String("caption-path", "./res/captions.conf", "файл с текстами постов.")
	proxyPath = flag.String("proxy-path", "./res/proxies.conf", "файл с проксями.")

	// Wipe flow settings.
	threads = flag.Uint64("t", 1, "кол-во потоков.")
	iters   = flag.Uint64("i", 1, "кол-во проходов.")
	timeout = flag.Uint64("T", 0, "перерыв между проходами (сек.)")

	// Additional settings.
	bufsize = flag.Uint64("buffer", 0, "размер буфера каналов.")
	limit   = flag.Uint64("limit", 1, "макс. число ошибок соединения для прокси перед удалением.")
	verbose = flag.Bool("v", false, "доп. логи для отладки.")
	domain  = flag.String("domain", "hk", "зеркало.\n\thk\n\tlife (depricated)")
	wait    = flag.Uint64("wait", 20, "ждём секунд печеньки")

	// Cloudflare init settings.
	initAtOnce = flag.Uint64("I", 1, "кол-во параллельно инициализируемых прокси.")
	sessions   = flag.Uint64("s", 1, "кол-во сессий на одну проксю (подробнее в документации).")
)

// If we want to use captions, but text initialization has failed.
var defaultCaptions = []string{
	"ALO YOBA ETO TI?",
	"NET, ON U BABUSHKI EST OLADUSHKI."}

type Mode struct {
	WipeMode    uint8
	AntiCaptcha uint8
	TextMode    uint8
}

type PostSettings struct {
	Sage         bool
	UseProxy     bool
	Colorize     bool
	FilesPerPost uint8
	Board        string
	Thread       uint64
}

type Content struct {
	Files    []File
	Proxies  []network.Proxy
	Captions []string
}

type WipeSettings struct {
	Threads uint64
	Iters   uint64
	Timeout uint64
}

type Env struct {
	Mode
	PostSettings
	WipeSettings
	*Content

	// Anti-captcha api key or passcode (yet not implemented)
	Key string

	Logger chan string        // Global synced logger.
	Filter chan network.Proxy // Global synced proxy filter.
	Status chan bool          // Global synced counter; True if post send, false if failed.

	// -v flag for verbose logging.
	Verbose bool

	// How many times proxy can fail HTTP request to captcha before get deleted.
	FailedConnectionsLimit uint64

	// -domain flag, depricated.
	Domain string

	// How many web driver instances will be spawned at once.
	InitAtOnce uint64

	// Cookie sessions for one proxy.
	Sessions uint64

	// Seconds to wait until cloudflare challenge will load.
	// Note -- this is not a timeout for the connection.
	// This one works when and only when connection was successful.
	WaitTime uint64
}

// Init all valid media files to env.Files. Should be called before initing captions.
func (env *Env) initEnvFiles(dir string) *Env {
	env.Files = make([]File, 0)
	if env.FilesPerPost != 0 {
		logger.Files.Log("инициализирую картинки...")
		cont, err := GetMedia(dir)
		if err == nil {
			env.Files = cont
			env.FilesPerPost = uint8(math.Min(float64(len(env.Files)), float64(env.FilesPerPost)))
			return env
		}
		logger.Files.Logf("ошибка инициализации: %v", err)
		logger.Files.Log("буду продолжать без использования файлов.")
		env.FilesPerPost = 0
	}
	return env
}

// Init all captions (post's texts) to env.Captions.
func (env *Env) initEnvCaptions(dir string) *Env {
	switch env.TextMode {
	case NO_CAPS:
		if env.FilesPerPost == 0 {
			logger.Captions.Log("ошибка, не могу постить без текста и без картинок.")
			os.Exit(1)
		}
		env.Captions = []string{""}
	case DEFAULT:
		logger.Captions.Log("буду использовать дефолтные тексты.")
		env.Captions = defaultCaptions
	case SCHIZO:
		logger.Captions.Log("SCHIZO not implemented yet")
		os.Exit(0)
	case FROM_POSTS:
		logger.Captions.Logf("получаю каталог тредов /%s/...", env.Board)
		caps, err := getPostsTexts(env.Board)
		if err != nil {
			logger.Captions.Logf("ошибка получения постов: %v", err)
			logger.Captions.Log("буду использовать дефолтные тексты.")
			env.Captions = defaultCaptions
			return env
		}
		env.Captions = caps
	case FROM_FILE:
		logger.Captions.Log("инициализирую тексты постов...")
		caps, err := GetCaptions(dir)
		if err == nil {
			env.Captions = caps
			logger.Captions.Logf("ok, %d текстов инициализировано.", len(caps))
			return env
		}
		logger.Captions.Log("ошибка инициализации, буду использовать дефолтные тексты.")
		env.Captions = defaultCaptions
	default:
		logger.Captions.Logf("неизвестный режим текста постов: %d, фатальная ошибка.", env.TextMode)
	}
	return env
}

// Check for validness and parse all proxies to env.Proxies.
func (env *Env) initEnvProxies(dir string) *Env {
	if env.UseProxy {
		logger.Proxies.Log("инициализирую прокси...")
		proxies, err := GetProxies(dir, int(env.Sessions))
		if err != nil {
			logger.Proxies.Logf("ошибка инициализации: %v", err)
			logger.Proxies.Log("не удалось инициализировать прокси, фатальная ошибка.")
			os.Exit(0)
		}
		env.Proxies = proxies
	}
	return env
}

// Parse all user input and return user environment struct.
func ParseEnv() (*Env, error) {
	flag.Parse()
	log.SetFlags(log.Ltime)

	env := Env{
		Mode: Mode{
			WipeMode:    uint8(*wipeMode),
			AntiCaptcha: uint8(*antiCaptcha),
			TextMode:    uint8(*textMode),
		},
		PostSettings: PostSettings{
			UseProxy:     *useProxy,
			Sage:         *useSage,
			Colorize:     *colorize,
			Thread:       *thread,
			Board:        *board,
			FilesPerPost: uint8(math.Min(float64(*files), 4)),
		},
		WipeSettings: WipeSettings{
			Threads: *threads,
			Iters:   *iters,
			Timeout: *timeout,
		},
		Content: new(Content),
		Key:     *antiCaptchaKey,
		Logger:  make(chan string, *bufsize),
		Filter:  make(chan network.Proxy, *bufsize),
		Status:  make(chan bool, *bufsize),

		FailedConnectionsLimit: *limit,
		Verbose:                *verbose,
		Domain:                 *domain,
		InitAtOnce:             *initAtOnce,
		Sessions:               *sessions,
		WaitTime:               *wait,
	}
	// Processing input errors.
	if env.InitAtOnce == 0 {
		return nil, fmt.Errorf("ошибка, -I должен быть больше нуля.")
	}
	if env.Sessions == 0 {
		return nil, fmt.Errorf("ошибка, -s должен быть больше нуля.")
	}
	if banned[env.Board] {
		return nil, fmt.Errorf("извини, но эту доску вайпать нельзя, она защищена магическим полем. Такие дела!")
	}
	if env.Domain == "life" {
		return nil, fmt.Errorf("2ch.life support deprecated")
	}
	if _, ok := domains[env.Domain]; !ok {
		return nil, fmt.Errorf("ошибка, не смогла распознать домен зеркала: %s", env.Domain)
	}
	if env.WipeMode == SINGLE && env.Thread == 0 {
		return nil, fmt.Errorf("ошибка, не указан ID треда.")
	}
	if env.WipeMode != SINGLE && env.Thread != 0 {
		return nil, fmt.Errorf("ошибка, ID треда указан, но режим не SingleThread.")
	}
	if env.AntiCaptcha != RUCAPTCHA && env.AntiCaptcha != OCR {
		return nil, fmt.Errorf("ошибка, пока доступны только OCR и RuCaptcha.")
	}

	env.initEnvFiles(*filesPath).
		initEnvCaptions(*capsPath).
		initEnvProxies(*proxyPath)

	if env.FilesPerPost == 0 && env.WipeMode == CREATING {
		return nil, fmt.Errorf("для создания тредов нужен хотя бы один файл!")
	}

	return &env, nil
}
