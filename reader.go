package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

func readFile(file string) [][2]float64 {
	dat := [][2]float64{}

	f, err := os.Open(file)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	reader := bufio.NewReader(f)

	for {
		l, err := reader.ReadString('\n')


		if err != nil {
			break
		}

		pointStrs := strings.Split(l, " ")

		p1, err := strconv.ParseFloat(pointStrs[0], 64)
		if err != nil {
			panic(err)
		}

		p2, err := strconv.ParseFloat(pointStrs[1], 64)
		if err != nil {
			panic(err)
		}

		point := [2]float64{
			p1, p2,
		}

		dat = append(dat, point)
	}

	return dat
}
