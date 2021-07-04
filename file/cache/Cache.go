package cache

import (
	"encoding/base64"
	"path/filepath"
	"strconv"
	"strings"
	defTime "time"

	"github.com/Wlademon/vkBot/time"

	fileB "github.com/Wlademon/vkBot/file"
)

const CACHE_SEP = "|"

var dirCache = ""

func InitCache(dir string) {
	dirCache = dir
}

type CacheFile struct {
	Key   string
	Value string
	TLL   defTime.Duration
}

func CreateForever(key string, value string) CacheFile {
	return Create(key, value, 0)
}

func Create(key string, value string, tll defTime.Duration) CacheFile {
	return CacheFile{
		Key:   key,
		Value: value,
		TLL:   tll,
	}
}

func createFileDir(key string) string {
	return strings.TrimRight(dirCache, string(filepath.Separator)) + string(filepath.Separator) + key + ".cache"
}

func (file CacheFile) Set() error {
	cacheF := fileB.File{Name: createFileDir(file.Key)}
	pref := "0"
	if file.TLL != 0 {
		pref = strconv.FormatInt(time.Now().Add(file.TLL).Unix(), 10)
	}
	_, err := cacheF.Write(pref + CACHE_SEP + base64.StdEncoding.EncodeToString([]byte(file.Value)))

	return err
}

func parseString(val string) (bool, string) {
	decodeString, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return false, ""
	}

	return true, string(decodeString)
}

func Get(key string) (bool, string) {
	cacheF := fileB.File{Name: createFileDir(key)}
	if !cacheF.IsExist() {
		return false, ""
	}
	lines, err := cacheF.ReadAllLines()
	if err != nil {
		return false, ""
	}
	cache := strings.Join(lines, "\n")
	durVal := strings.Split(cache, CACHE_SEP)
	if durVal[0] == "0" {
		return parseString(durVal[1])
	}
	parseInt, err := strconv.ParseInt(durVal[0], 10, 64)
	if err != nil {
		return false, ""
	}
	if parseInt >= time.Now().Unix() {
		return parseString(durVal[1])
	}
	cacheF.Delete()

	return false, ""
}

func Flush(key string) {
	cacheF := fileB.File{Name: createFileDir(key)}
	cacheF.Delete()
}
