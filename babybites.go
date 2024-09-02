package main 
import (
	"fmt"
	"bufio"
	"regexp"
	"os"
	"strings"
	"strconv"
)

func main() {
	var actual int
	var inStr string 
	fmt.Scanln(&actual)
	reader := bufio.NewReader(os.Stdin)
    inStr, _ = reader.ReadString('\n')

	re := regexp.MustCompile("\\d+|\\w+")

	words := re.FindAllString(inStr, -1)

	count := 0
	for _, word := range words {
		count++
		if(!strings.Contains(word, "mumble")) {
			cunt, _ := strconv.Atoi(word)
			if(cunt != count) {
				fmt.Println("something is fishy")
				os.Exit(0)
			}
		} 

	}
	fmt.Println("makes sense")
}