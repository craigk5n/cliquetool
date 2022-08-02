package main

// GroupParamters are used to specify different group density requirements.
// A node "density" defines how connected a node is to all the other nodes in
// the group.  For example, if Node A is connected to 8 of the other 10 nodes
// within a 11-node group, then Node A has a density of 0.8 within the group.
// Within the group paramater file, we define the minimum density required for
// a node to be considered part of the group.

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Param holds settings for a group of a specific size
type Param struct {
	groupSize int
	// minimum density for a Node to be part of a group
	density float64
	// minimum number of links (calculated from density and groupSize above)
	minLinks int
}

// GroupParameters holds settings for all group sizes
type GroupParameters struct {
	filename string
	params   []Param
}

// Read the group parameter file, which defines the settings for when to merge
// a group of a specific size.
func readParamFile(filename string) (GroupParameters, error) {
	var pt GroupParameters
	pt.filename = filename
	pt.params = []Param{}
	fp, err := os.Open(filename)
	if err != nil {
		return pt, err
	}
	defer fp.Close()
	fileScanner := bufio.NewScanner(fp)
	lineNo := 0
	for fileScanner.Scan() {
		var p Param
		line := fileScanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		lineNo++
		fields := strings.Fields(line)
		//fmt.Printf("Line %d: %s %s\n", lineNo, fields[0], fields[1])
		if len(fields) != 2 {
			return pt, fmt.Errorf("invalid data on line %d", lineNo)
		}
		p.groupSize, _ = strconv.Atoi(fields[0])
		p.density, _ = strconv.ParseFloat(fields[1], 32)
		if p.density > 1.0 {
			// This is actually min links rather than density
			p.density = p.density / float64(p.groupSize)
		}
		for len(pt.params) < p.groupSize {
			existingLen := len(pt.params)
			var p2 Param = Param{existingLen, 1.0, existingLen - 1}
			pt.params = append(pt.params, p2)
		}
		// Calculate the minimum number of links for each group size
		// so we don't have to keep doing it over and over later.
		p.minLinks = int(math.Ceil(p.density * float64(p.groupSize)))
		if p.minLinks >= p.groupSize {
			p.minLinks = p.groupSize - 1
		}
		//fmt.Printf("Line %d: groupSize=%v density=%v minLinks=%v\n", lineNo, p.groupSize, p.density, p.minLinks)
		pt.params = append(pt.params, p)
		//fmt.Printf("Now have %d lines of params\n", len(pt.params))
	}
	return pt, nil
}

func minLinksForGroupSize(paramTable *GroupParameters, groupSize int) int {
	// If no params, we are searching for complete/dense groups with density = 100%
	if paramTable == nil {
		return groupSize - 1
	}
	if groupSize >= len(paramTable.params)-1 {
		groupSize = len(paramTable.params) - 1
	}

	return paramTable.params[groupSize].minLinks
}

func minDensityForGroupSize(paramTable *GroupParameters, groupSize int) float64 {
	// If no params, we are searching for complete/dense groups with density = 100%
	if paramTable == nil {
		return 1.0
	}
	if groupSize >= len(paramTable.params) {
		groupSize = len(paramTable.params)
	}

	return paramTable.params[groupSize].density
}
