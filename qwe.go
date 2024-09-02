package main 
import (
	"fmt"
	"regexp"
	"strconv"
	"os"
)

func main() {
	suitToIndexMap := map[byte]int{'P' : 0, 'K' : 1, 'H' : 2, 'T' : 3}
	indexToSuitMap := map[int]byte{0 : 'P', 1 : 'K', 2: 'H', 3 : 'T'}
    suitCountMap := map[byte]int{'P' : 0, 'H' : 0, 'K' : 0, 'T' : 0}
	var suitDuplicateArr [4][13]bool
	re := regexp.MustCompile("\\w\\d\\d")

	var input string 
	fmt.Scanln(&input)

	matches := re.FindAllString(input, -1)

	for _, match := range matches {
		suitCountMap[match[0]]++
		num, _ := strconv.Atoi(match[1:])
		if(suitDuplicateArr[suitToIndexMap[match[0]]][num]) {
			fmt.Println("GRESKA");
			os.Exit(69)
		} else {
			suitDuplicateArr[suitToIndexMap[match[0]]][num] = true
		}
	}

	for i:=0; i < 4; i++ {
		fmt.Printf("%d ", 13 - suitCountMap[indexToSuitMap[i]])
	}
	fmt.Println()
}