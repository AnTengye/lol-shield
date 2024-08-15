package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
)

func main() {
	if !admin.IsAdmin() {
		log.Println("需要管理员权限，请以管理员身份运行")
		return
	}

	log.Println("正在以管理员权限运行")
	v3, s, err := lcu.GetLcuToken(false)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = os.WriteFile("lcu.token", []byte(fmt.Sprintf("%d %s", v3, s)), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
