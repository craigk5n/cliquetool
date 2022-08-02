package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Functions for handling nodes (also called vertices in graph theory).
// Note that we only consider an undirected graph.

func trimQuotes(str string) string {
	if str[0] == '"' {
		str = str[1:]
	}
	if str[len(str)-1] == '"' {
		str = str[:len(str)-1]
	}
	return str
}

// Read a data file in CSV format where each line is just two node
// identifiers separated by a comma.
func readNodeFile(inputFile string, verbose int) (NodeDb, error) {
	numInvalid := 0
	numValid := 0

	// init empty slice of Node
	nodeDb := initNodeDb()

	fp, err := os.Open(inputFile)
	if err != nil {
		return nodeDb, err
	}
	defer fp.Close()
	fileScanner := bufio.NewScanner(fp)
	lineNo := 0
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineNo++
		arr := strings.Split(strings.TrimSpace(line), ",")
		if len(arr) != 2 {
			numInvalid++
		} else {
			numValid++
			name1 := trimQuotes(arr[0])
			// Have we seen this external Id before?
			id1, _ := getNodeIDForExternalName(nodeDb, name1)
			if id1 < 0 {
				// First time for external Id.
				id1 = createNode(&nodeDb, name1)
				debugMessage(fmt.Sprintf("Created new Node name='%s' with id=%d", name1, id1))
			}
			name2 := trimQuotes(arr[1])
			id2, _ := getNodeIDForExternalName(nodeDb, name2)
			if id2 < 0 {
				id2 = createNode(&nodeDb, name2)
				debugMessage(fmt.Sprintf("Created new Node name='%s' with id=%d", name2, id2))
			}
			addLink(&nodeDb, id1, id2)
		}
	}

	debugMessage(fmt.Sprintf("Read %d lines: %d valid, %d invalid",
		lineNo, numValid, numInvalid))

	return nodeDb, nil
}

// Read a data file in DIMACs ASCII format.
func readNodeFileDimacs(inputFile string, verbose int) (NodeDb, error) {
	numInvalid := 0
	numValid := 0

	// init empty slice of Node
	nodeDb := initNodeDb()

	fp, err := os.Open(inputFile)
	if err != nil {
		return nodeDb, err
	}
	defer fp.Close()
	fileScanner := bufio.NewScanner(fp)
	lineNo := 0
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineNo++
		if strings.HasPrefix(line, "e ") {
			arr := strings.Split(strings.TrimSpace(line), " ")
			if len(arr) != 3 {
				numInvalid++
			} else {
				numValid++
				name1 := trimQuotes(arr[1])
				// Have we seen this external Id before?
				id1, _ := getNodeIDForExternalName(nodeDb, name1)
				if id1 < 0 {
					// First time for external Id.
					id1 = createNode(&nodeDb, name1)
					debugMessage(fmt.Sprintf("Created new Node name='%s' with id=%d", name1, id1))
				}
				name2 := trimQuotes(arr[2])
				id2, _ := getNodeIDForExternalName(nodeDb, name2)
				if id2 < 0 {
					id2 = createNode(&nodeDb, name2)
					debugMessage(fmt.Sprintf("Created new Node name='%s' with id=%d", name2, id2))
				}
				addLink(&nodeDb, id1, id2)
			}
		}
	}

	debugMessage(fmt.Sprintf("Read %d lines: %d valid, %d invalid",
		lineNo, numValid, numInvalid))

	return nodeDb, nil
}

func createGroupFromNodes(name string, ids []int) *Group {
	var group Group
	group.groupName = name
	group.ids = make([]int, len(ids))
	copy(group.ids, ids)
	return &group
}

// Load a Group CSV file typically written with this tool's -writecsv
// parameter.
func loadGroupDbFromCSV(path string, nodeDb *NodeDb) (*GroupDb, error) {
	var groupDb GroupDb
	groupDb.groups = make([]Group, 0)

	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	fileScanner := bufio.NewScanner(fp)
	lineNo := 0
	lastGroupName := ""
	nodes := make([]int, 0)
	numInvalid := 0
	numValid := 0
	for fileScanner.Scan() {
		line := fileScanner.Text()
		lineNo++
		// Ignore header and blank lines
		if len(line) > 0 && !strings.HasPrefix(line, "group,member") {
			arr := strings.Split(strings.TrimSpace(line), ",")
			if len(arr) != 2 {
				numInvalid++
			} else {
				numValid++
				groupName := trimQuotes(arr[0])
				// Have we seen this external Id before?
				name := arr[1]
				id, _ := getNodeIDForExternalName(*nodeDb, arr[1])
				if id < 0 {
					// Unknown ID.  This gruop file might not match the loaded NodeDb.
					fmt.Printf("Warning: Unknown ID '%s' not found in NodeDb\n", name)
					debugMessage(fmt.Sprintf("Warning: Unknown ID '%s' not found in NodeDb", name))
				}
				if groupName != lastGroupName {
					// changing groups... add last one to GroupDb
					if len(nodes) > 0 {
						group := createGroupFromNodes(groupName, nodes)
						addGroup(nodeDb, &groupDb, group, false)
					}
					nodes = make([]int, 1)
					nodes[0] = id
					lastGroupName = groupName
				} else {
					nodes = append(nodes, id)
				}
			}
		}
	}
	if len(lastGroupName) > 0 {
		// Add last group
		group := createGroupFromNodes(lastGroupName, nodes)
		addGroup(nodeDb, &groupDb, group, false)
	}
	fmt.Printf("Loaded %d groups from %s\n", len(groupDb.groups), path)
	return &groupDb, nil
}
