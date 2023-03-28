# kuklowipe
Дочка [вот этого чуда](https://github.com/neuroliptica/traumatic), переписанная с учётом ошибок своего предка. Не работает по причине клауда.

## Имеет
- Простенький графический интерфейс.
- Поддержка нового движка абучана.
- Поддержка многопотока.
- Поддержка всех видов прокси.
- Несколько режимов вайпа (вся доска, один тред, создание тредов).
- Онлайн сервисы решения капчи (RuCaptcha).
- Поддержка всех основных видов файлов (jpg, png, jpeg, mp4, webm, gif).
- Цветовые маски для картинок по желанию.
- Несколько режимов постов (без текста, брать из файла, копировать из других тредов).
- Ручная настройка без GUI всего при помощи флагов.

## Ожидается
- (!) Нормальный обход клауды.
- Вайп с пасскодом.
- Остальные антикапчи (XCaptcha, AntiCaptcha).
- Распознавание капчи (OCR [сильно под вопросом])

## Собрать
Надо Go >= 1.18
```bash
$ git clone https://github.com/neuroliptica/dollwipe.git
$ cd dollwipe
$ go build
```

## Собрать GUI
Надо Vala + GTK3
```bash
$ cd dollwipe/gui
$ valac -o dollgui --Xcc="-I./" gui.vala fails.vala utils.vala base.vala consts.vala --pkg gtk+-3.0 --pkg posix
```

## License
MIT

## Ошибки
В [issues гитхаба](https://github.com/neuroliptica/dollwipe/issues) или мне в [телеграм](https://t.me/seharehare).
