package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/gin-gonic/gin"
)

var (
	secret string
	cmd    string
	port   int
)

func genSha1(data string, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(data))
	return "sha1=" + hex.EncodeToString(h.Sum(nil))
}

func runCommand() error {
	excmd := exec.Command("/bin/sh", "-c", cmd)
	_, err := excmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}

	return err
}

func init() {
	flag.StringVar(&secret, "secret", "", "Github webhook secret")
	flag.StringVar(&cmd, "cmd", "", "Exec command")
	flag.IntVar(&port, "port", 8833, "Port")
	flag.Parse()

	log.SetFlags(log.Lshortfile | log.LstdFlags)
	gin.SetMode(gin.ReleaseMode)

}

func main() {
	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		event := c.GetHeader("X-GitHub-Event")
		if event == "push" {
			signature := c.GetHeader("X-Hub-Signature")
			data, _ := io.ReadAll(c.Request.Body)

			if signature == genSha1(string(data), secret) {
				err := runCommand()
				if err != nil {
					c.JSON(500, gin.H{"message": err.Error()})
				} else {
					c.JSON(200, gin.H{})
				}
			} else {
				c.JSON(403, gin.H{
					"message": "Secret verification failed",
				})
			}
		}
	})

	fmt.Printf("GWHOOK is runing on 0.0.0.0:%v\n", port)
	r.Run(fmt.Sprintf(":%v", port))
}
