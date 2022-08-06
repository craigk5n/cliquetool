package main

// The are different phases that this tool will normally run.
// 1 - Find: Find all cliques, where a clique is a fully connected graph
//     where all nodes are connected to all other nodes in the graph.
// 2 - Build: Build upon the cliques found in step 1 to enlarge the groups
//     to include nodes that are highly connected to the other nodes but are
//     not fully connected using the parameter settings to determine what
//     percentage of existing nodes a potential new node must be connected to.
//     This process will result in many overlapping groups.
// 3 - Merge: Merge overlapping groups.  This is done by the percent overlap
//     setting and the resulting groups may violate the rules in the
//     paramater settings.
// 4 - Prune: Prune groups down by kicking out group members that violate the
//     parameter settings.
// There are other functions that can also be run in addition to the phases:
//   Sort: sort the internal database of groups by placing the largest groups
//     in front
//   Rename: rename the internal groups so that the groups at the beginning
//     of the list is 1, then 2, etc.

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CliqueToolVersion represents the release number
const CliqueToolVersion string = "1.0.0"

var verboseLevel int = 0

func fatalError(message string) {
	_, _ = os.Stderr.WriteString(message + "\n")
	os.Exit(1)
}

func debugMessage(message string) {
	if verboseLevel > 0 {
		fmt.Printf("%s\n", message)
	}
}

func main() {
	//paramFile := nil
	inputFile := ""
	minGroupSize := 3
	searchPasses := 1
	//outputFile := ""
	//outputDir := "."
	var nodeDb NodeDb
	var err error
	var groupDb *GroupDb
	needWrite := false
	var groupParams GroupParameters
	mergeOverlapRatio := DefaultMergeRatioThreshold
	mergeMaxMissingGroupLinks := -1
	groupPrefix := "cliquetool"
	reportFormat := ReportFormatMarkdown
	reportSortBySize := true
	numThreads := 1

	// Rather than using the built-in command line parsing, we handle them one at a time here.
	// This is because the order of the parameters matters.
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		//fmt.Printf("Arg %d: %s\n", i, arg)
		if arg == "-v" || arg == "-verbose" {
			verboseLevel++
			logInit("cliquetool.log", verboseLevel)
		} else if arg == "-version" {
			fmt.Printf("cliquetool version %s\n", CliqueToolVersion)
			os.Exit(0)
		} else if arg == "-minsize" {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -minsize requres an integer parameter")
			}
			i++
			minGroupSize, err = strconv.Atoi(os.Args[i])
			if err != nil {
				fatalError("Invalid parameter for -minGroupSize")
			}
		} else if arg == "-t" || arg == "-threads" {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -threads requres an integer parameter")
			}
			i++
			numThreads, err = strconv.Atoi(os.Args[i])
			if err != nil {
				fatalError("Invalid parameter for -threads")
			}
		} else if arg == "-prefix" {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -prefix requres an string parameter")
			}
			i++
			groupPrefix = os.Args[i]
		} else if arg == "-passes" {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -passes requres an integer parameter")
			}
			i++
			searchPasses, err = strconv.Atoi(os.Args[i])
			if err != nil {
				fatalError("Invalid parameter for -passes")
			}
		} else if arg == "-i" || arg == "-infile" {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -infile requres a filename")
			}
			if len(inputFile) > 0 {
				fatalError("Error: only one input file can be specified\n")
			}
			i++
			inputFile = os.Args[i]
			nodeDb, err = readNodeFile(inputFile, verboseLevel)
			if err != nil {
				fatalError("Error opening file: " + inputFile)
			}
			if verboseLevel > 0 {
				fmt.Printf("Found %d unique nodes\n", len(nodeDb.nodes))
				if verboseLevel > 1 {
					dumpDb(nodeDb)
				}
			}
			fmt.Printf("Loaded %d nodes from %s.\n", len(nodeDb.nodes), inputFile)
		} else if strings.HasPrefix(arg, "-dimacs") {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -dimacsinfile requres a filename")
			}
			if len(inputFile) > 0 {
				fatalError("Error: only one input file can be specified\n")
			}
			i++
			inputFile = os.Args[i]
			nodeDb, err = readNodeFileDimacs(inputFile, verboseLevel)
			if err != nil {
				fatalError("Error opening file: " + inputFile)
			}
			if verboseLevel > 0 {
				fmt.Printf("Found %d unique nodes\n", len(nodeDb.nodes))
				if verboseLevel > 1 {
					dumpDb(nodeDb)
				}
			}
			fmt.Printf("Loaded %d nodes from %s.\n", len(nodeDb.nodes), inputFile)
		} else if arg == "-loadgroupcsv" {
			// Must have already loaded data with -infile
			if len(nodeDb.nodes) == 0 {
				fatalError("You must load data with -infile or -dimacsinfile before using -loadgroupcsv")
			}
			i++
			inputFile = os.Args[i]
			if groupDb != nil && len(groupDb.groups) > 0 {
				fmt.Printf("Warning: discarding previously found %d groups\n", len(groupDb.groups))
			}
			groupDb, err = loadGroupDbFromCSV(inputFile, &nodeDb)
			if err != nil {
				fatalError(fmt.Sprintf("Error loading groups from file: %v", err))
			}
		} else if arg == "-f" || arg == "-find" {
			// Must have already loaded data with -infile
			if len(nodeDb.nodes) == 0 {
				fatalError("You must load data with -infile or -dimacsinfile before using -find")
			}
			groupDb, err = findGroups(&nodeDb, minGroupSize, searchPasses, verboseLevel, numThreads)
			if err != nil {
				fatalError(fmt.Sprintf("Error finding groups: %v", err))
			}
			needWrite = true
		} else if arg == "-p" || strings.HasPrefix(arg, "-param") {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -param requres a filename")
			}
			if len(groupParams.filename) > 0 {
				fatalError("Error: only one group parameter file can be specified\n")
			}
			i++
			paramFile := os.Args[i]
			groupParams, err = readParamFile(paramFile)
			if err != nil {
				fatalError(fmt.Sprintf("Error reading group parameter file: %v", err))
			}
		} else if arg == "-b" || strings.HasPrefix(arg, "-build") {
			// Must have already loaded data with -infile and specified output with -outfile
			if len(nodeDb.nodes) == 0 {
				fatalError("You must load node data with -infile before using -build")
			}
			if groupDb == nil || len(groupDb.groups) == 0 {
				fatalError("You must load group data with -loadcsv or using -infile and -find before using -build")
			}
			if len(groupParams.params) == 0 {
				fatalError("You must load a group paramater file with -param before using -build")
			}
			groupDb2 := buildGroups(&nodeDb, groupDb, groupParams, false, minGroupSize, verboseLevel)
			groupDb = groupDb2
			needWrite = true
		} else if arg == "-rename" {
			renameGroups(groupDb, groupPrefix)
			needWrite = true
		} else if strings.HasPrefix(arg, "-mergeparam") {
			if i+1 >= len(os.Args) {
				fatalError("Parameter -mergeparameter requres a decimal")
			}
			i++
			mergeOverlapRatio, err = strconv.ParseFloat(os.Args[i], 64)
			if err != nil {
				fatalError(fmt.Sprintf("Invalid mergeparameter: %v", err))
			}
		} else if arg == "-merge" {
			groupDb = mergeGroups(&nodeDb, groupDb, mergeOverlapRatio,
				mergeMaxMissingGroupLinks, &groupParams, minGroupSize, verboseLevel)
			needWrite = true
		} else if arg == "-removesubsets" {
			removeGroupSubsets(groupDb, verboseLevel)
		} else if arg == "-prune" {
			// Must have already loaded data with -infile and specified output with -outfile
			if len(nodeDb.nodes) == 0 {
				fatalError("You must load node data with -infile before using -prune")
			}
			if groupDb == nil || len(groupDb.groups) == 0 {
				fatalError("You must load group data with -loadcsv or using -infile and -find before using -prune")
			}
			if len(groupParams.params) == 0 {
				fatalError("You must load a group paramater file with -param before using -prune")
			}
			pruneGroups(&nodeDb, groupDb, &groupParams, minGroupSize, verboseLevel)
			removeGroupSubsets(groupDb, verboseLevel)
			needWrite = true
		} else if arg == "-sort" {
			sortGroupDb(groupDb)
		} else if arg == "-writecsv" {
			// Write found groups to a file.  Make sure we have groups.
			if groupDb == nil || len(groupDb.groups) == 0 {
				fatalError("You must find or load groups before using -writecsv")
			}
			if i+1 >= len(os.Args) {
				fatalError("Parameter -writecsv requres a filename")
			}
			i++
			outputFile := os.Args[i]
			err = writeGroupFile(&nodeDb, groupDb, outputFile, minGroupSize, verboseLevel)
			if err != nil {
				fmt.Printf("Error writing file %s: %s\n", outputFile, err)
			}
			needWrite = false
		} else if arg == "-html" {
			// For use with -verify
			reportFormat = ReportFormatHtml
		} else if arg == "-markdown" {
			// For use with -verify
			reportFormat = ReportFormatMarkdown
		} else if arg == "-reportbysize" {
			// Sort groups in report by group size, largest first
			reportSortBySize = true
		} else if arg == "-reportbyconcern" {
			// Sort groups in report with those most likely missing some nodes first
			reportSortBySize = false
		} else if arg == "-verify" || arg == "-report" {
			if groupDb == nil || len(groupDb.groups) == 0 {
				fatalError("You must find or load groups before using -verify or -report")
			}
			if i+1 >= len(os.Args) {
				fatalError("Parameters -verify and -report requre a filename")
			}
			i++
			outputFile := os.Args[i]
			if outputFile == "-" {
				outputFile = "" // Allow stdout when "-" is passed as output file
			}
			err = verifyGroupDb(&nodeDb, groupDb, outputFile, reportFormat, reportSortBySize)
			if err != nil {
				fmt.Printf("Error writing file %s: %s\n", outputFile, err)
			}
		} else {
			fatalError("Unrecognized parameter: " + os.Args[i])
		}
	}
	if needWrite {
		fmt.Println("Warning: final data was not saved to a file.")
	}
}
