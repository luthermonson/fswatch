# fswatch

Simple file system watcher to tracks create/write/remove events on a specific directory and echo them out so you can see changes. Inspired by [Go Watcher](https://github.com/mattdamon108/gw) and relying heavily on the [fsnotify](https://github.com/fsnotify/fsnotify) package.

#  Usage

Will watch the current directory by default when you run `fswatch` and to specify what directory pass as the first argument `fswatch /tmp` 

