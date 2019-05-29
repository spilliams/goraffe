package guru

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Call(mode string, position string) error {
	cmdStr := fmt.Sprintf("guru -json %s %s", mode, position)
	logrus.Debugf(cmdStr)
	worker := exec.Command("bash", "-c", cmdStr)
	worker.Stdout = os.Stdout
	worker.Stdin = os.Stdin
	worker.Stderr = os.Stderr
	return worker.Run()
}
