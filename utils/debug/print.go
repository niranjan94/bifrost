package debug

import (
	"bufio"
	"github.com/niranjan94/bifrost/utils"
	"github.com/sirupsen/logrus"
	"strings"
)

// PrintMultilineOutput takes in a multiline input and prints it line by line using the project's logger.
func PrintMultilineOutput(input string)  {
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		logrus.Debug(utils.ToValidUTF8(scanner.Text()))
	}
}
