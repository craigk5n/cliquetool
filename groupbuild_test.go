package main

import (
	"testing"
)

func TestSortPotentials(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	// Create two fully connected groups of size 10
	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	group1 := &groupDb.groups[0]
	//group2 := &groupDb.groups[1]

	// Ask if there are any nodes that link to 5 nodes in group1 (no)
	potentials := nodesThatLinkToGroup(&nodeDb, group1, 5)
	if len(potentials) != 0 {
		t.Errorf("Expected potentials of 0, %d", len(potentials))
	}
	// Connect new node to first 5 nodes of group1
	newNodeID := createNewNodeAndConnectToGroup(&nodeDb, &groupDb,
		group1, "group1node_2_", 5)
	potentials = nodesThatLinkToGroup(&nodeDb, group1, 6)
	if len(potentials) != 0 {
		t.Errorf("Expected potentials of 0, %d", len(potentials))
	}
	potentials = nodesThatLinkToGroup(&nodeDb, group1, 5)
	if len(potentials) != 1 || potentials[0] != newNodeID {
		t.Errorf("Expected potentials of 1, %d", len(potentials))
	}
	// Connect another new node to first 7 nodes of group1
	newNodeID2 := createNewNodeAndConnectToGroup(&nodeDb, &groupDb,
		group1, "group1node_3_", 7)
	potentials = nodesThatLinkToGroup(&nodeDb, group1, 6)
	if len(potentials) != 1 {
		t.Errorf("Expected potentials of 1, not %d", len(potentials))
	}
	potentials = nodesThatLinkToGroup(&nodeDb, group1, 5)
	if len(potentials) != 2 || potentials[0] != newNodeID2 {
		t.Errorf("Expected potentials of 2, not %d", len(potentials))
	}
}

func TestExpandGroup(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	testfile := "testdata/test_params.txt"
	groupParams, _ := readParamFile(testfile)

	// Create two fully connected groups of size 10
	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	group1 := &groupDb.groups[0]
	//group2 := &groupDb.groups[1]

	// Connect new node to first 5 nodes of group1
	createNewNodeAndConnectToGroup(&nodeDb, &groupDb,
		group1, "group1node_2_", 5)
	// Connect another new node to first 8 nodes of group1
	createNewNodeAndConnectToGroup(&nodeDb, &groupDb,
		group1, "group1node_3_", 8)
	beforeCnt := len(groupDb.groups)
	// Note: expandGroup does not add nodes to existing groups.
	// It clones and adds to existing groups with the new node then
	// we later delete subsets.
	ret := expandGroup(&nodeDb, &groupDb, group1, &groupParams, false, 5)
	if !ret {
		t.Errorf("Expected expandGroup to return true")
	}
	afterCnt := len(groupDb.groups)
	if afterCnt-beforeCnt != 1 {
		t.Errorf("Expected expandGroup to add 1 group")

	}
	lastGroup := groupDb.groups[len(groupDb.groups)-1]
	//dumpGroups(&nodeDb, &groupDb)
	if len(lastGroup.ids) != 11 {
		t.Errorf("Expected group size of 11, not %d", len(group1.ids))
	}
}
