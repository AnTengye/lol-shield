package lcu

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/AnTengye/lol-shield/internal/pkg/windows/process"
	"github.com/pkg/errors"
)

const (
	lolUxProcessName = "LeagueClientUx.exe"
)

var (
	lolCommandlineReg     = regexp.MustCompile(`--remoting-auth-token=(.+?)" "--app-port=(\d+)"`)
	ErrLolProcessNotFound = errors.New("未找到lol进程")
)

func GetLcuToken(debug bool) (port int, token string, err error) {
	if debug {
		return getLcuTokenFromFile()
	}
	cmdline, err := process.GetProcessCommand(lolUxProcessName)
	if err != nil {
		err = ErrLolProcessNotFound
		return
	}
	btsChunk := lolCommandlineReg.FindSubmatch([]byte(cmdline))
	if len(btsChunk) < 3 {
		return port, token, ErrLolProcessNotFound
	}
	token = string(btsChunk[1])
	port, err = strconv.Atoi(string(btsChunk[2]))
	return
}

func getLcuTokenFromFile() (port int, token string, err error) {
	file, err := os.ReadFile("lcu.token")
	if err != nil {
		return 0, "", err
	}
	split := strings.Split(string(file), " ")
	if len(split) != 2 {
		return 0, "", errors.New("invalid lcu.token file")
	}
	port, _ = strconv.Atoi(split[0])
	token = split[1]
	_ = os.Remove("lcu.token")
	return port, token, nil
}
