package env

import (
	"io/fs"
	"os"
	"os/user"
	"path/filepath"

	"tractor.dev/toolkit/engine"
	"tractor.dev/toolkit/exp/vfs/watchfs"
)

// TODO:
// - is root? need more usecases
// - platform? GOOS+GOARCH or something else?
// - is tty? platform specific impl
// - modes:
// 		dev: check if
//			running with go run or tractor dev
//			or if in source dir. other heuristics?
//		bundle: check if running as app bundle (lives in .app)
//		standalone: not in dev mode, not in bundle mode
// - more directories as fs.FS:
// 		Data/Config: os.UserConfigDir
//		Home: os.UserHomeDir
//		Cache: os.UserCacheDir
//		Temp: os.TempDir
//		Desktop: platform dependent
//		Downloads: platform dependent
//		DotConfig: $HOME/.<whatever>
//		Bundle: the app bundle path
//		should we determine a folder within these to always use?
//		this would require a way to determine and let user change
//		the app "name" like bundle ID

func init() {
	var err error

	// we're going to panic until we know the cases
	// we'd expect an error and its ok

	if ConfigDir, err = os.UserConfigDir(); err != nil {
		panic(err)
	}

	if CurrentPath, err = os.Getwd(); err != nil {
		panic(err)
	}
	CurrentDir = FS(CurrentPath)

	if CurrentUser, err = user.Current(); err != nil {
		panic(err)
	}

	if HomePath, err = os.UserHomeDir(); err != nil {
		panic(err)
	}

	if Hostname, err = os.Hostname(); err != nil {
		panic(err)
	}
}

// Var returns an environment variable by name or
// a default value if the variable is empty.
func Var(v ...string) string {
	vv := os.Getenv(v[0])
	if vv == "" && len(v) > 1 {
		return v[1]
	}
	return vv
}

// func Get[T string | bool](v ...string) (out T) {
// 	val := os.Getenv(v[0])
// 	if val == "" && len(v) > 1 {
// 		val = v[1]
// 	}
// 	switch reflect.TypeOf(out).Kind() {
// 	case reflect.Bool:
// 		if val == "false" || val == "0" {
// 			return false
// 		}
// 		return val != ""
// 	default:
// 		return val.(T)
// 	}
// }

func DataPath(subdir ...string) string {
	path := filepath.Join(ConfigDir, engine.Identifier, filepath.Join(subdir...))
	os.MkdirAll(filepath.Dir(path), 0766)
	return path
}

func FS(path string) fs.FS {
	// TODO: optimize by keeping single root FS, return subset
	return watchfs.New(os.DirFS(path))
}

var (
	ConfigDir   string
	CurrentPath string
	CurrentDir  fs.FS
	CurrentUser *user.User

	HomePath string

	Hostname string
)
