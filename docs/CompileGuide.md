# Compile Guide
###### зависимости: `Go` `Vala` `GTK-3`

## Linux
Установка зависимостей (пример для Debian):

```bash
$ sudo apt-get update
$ sudo apt-get install golang-go libgtk-3-0 libgtk-3-dev valac
```

Примечание: в репах может быть не самая свежая версия Go. Если в связи с этим возникает проблема с компиляцией, то ставьте напрямую.

Собственно сборка:

```bash
$ git clone https://github.com/neuroliptica/dollwipe.git
$ cd dollwipe
$ make build
```

## Windows
Для начала следует поставить [msys2](https://www.msys2.org). Далее все пакеты устанавливаются для 64х-битной подсистемы (название пакета включает mingw_w64-x86_64...)

Компилятор Go для винды ставьте отдельно и добавляйте его в PATH. Установка остальных зависимостей:

```powershell 
> pacman -Syu
> pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-vala mingw-w64-x86_64-make
```

Далее добавьте полный путь до `msys2/mingw64/bin` в PATH. Собственно сборка:

```powershell 
> git clone https://github.com/neuroliptica/dollwipe.git
> cd dollwipe
> make build
```
