package cmd

import (
	"fmt"
)

const (
	colorBlue  = "\x1b[38;2;0;85;212m"
	colorRed   = "\x1b[38;2;255;42;42m"
	colorWhite = "\x1b[38;2;255;255;255m\x1b[1m"
	colorReset = "\x1b[0m"
)

func PrintAscii() {
	lines := []string{
		"",
		"                    " + colorBlue + "*#########################" + colorReset,
		"                   " + colorBlue + "###########################" + colorReset,
		"                  " + colorBlue + "############################" + colorReset,
		"                  " + colorBlue + "#######" + colorReset,
		"      " + colorWhite + "@@@    @@@." + colorReset + " " + colorBlue + "######      ##" + colorReset + "  " + colorWhite + "@@@.    @@@" + colorReset,
		"      " + colorWhite + "@@@    @@@." + colorReset + " " + colorBlue + "####        ##" + colorReset + "  " + colorWhite + "@@@@   @@@@" + colorReset,
		"      " + colorWhite + "@@@@@@@@@@." + colorReset + " " + colorBlue + "####" + colorReset + " " + colorRed + "****" + colorReset + "  " + colorBlue + "###" + colorReset + "  " + colorWhite + "@@@@@ @@@@@" + colorReset,
		"      " + colorWhite + "@@@@@@@@@@." + colorReset + " " + colorBlue + "###" + colorReset + "  " + colorRed + "****" + colorReset + " " + colorBlue + "####" + colorReset + "  " + colorWhite + "@@@@@*@@@@@." + colorReset,
		"      " + colorWhite + "@@@    @@@." + colorReset + " " + colorBlue + "##       :####" + colorReset + "  " + colorWhite + "@@ @@@@@ @@@" + colorReset,
		"      " + colorWhite + "@@@    @@@." + colorReset + " " + colorBlue + "#+      ######" + colorReset + "  " + colorWhite + "@@  @@@  @@@" + colorReset,
		"                         " + colorBlue + "######*" + colorReset,
		"    " + colorBlue + "###########################=" + colorReset,
		"    " + colorBlue + "###########################" + colorReset,
		"    " + colorBlue + "#########################+" + colorReset,
	}

	for _, line := range lines {
		fmt.Println(line)
	}
}
