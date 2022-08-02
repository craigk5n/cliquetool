package main

import (
	"fmt"
	"testing"
)

func createConnectedNodes(nodeDb *NodeDb, groupDb *GroupDb, size int, namePrefix string) {
	var names []string = []string{}
	// Create nodes
	for i := 0; i < size; i++ {
		name := namePrefix + fmt.Sprintf("%03d", i)
		names = append(names, name)
		createNode(nodeDb, name)
	}
	// Connect them
	for i := 0; i < size; i++ {
		nodeID1, _ := getNodeIDForExternalName(*nodeDb, names[i])
		for j := i + 1; j < size; j++ {
			nodeID2, _ := getNodeIDForExternalName(*nodeDb, names[j])
			addLink(nodeDb, nodeID1, nodeID2)
		}
	}
	// Add all to a group
	var group Group
	group.groupName = generateGroupName()
	group.ids = []int{}
	group.linkCounts = make([]int, size) // will populate later in updateGroupLinkCounts
	for i := 0; i < size; i++ {
		nodeID1, _ := getNodeIDForExternalName(*nodeDb, names[i])
		group.ids = append(group.ids, nodeID1)
	}
	addGroup(nodeDb, groupDb, &group, false)
	updateGroupLinkCounts(nodeDb, &group)
}

func createNewNodeAndConnectToGroup(nodeDb *NodeDb, groupDb *GroupDb, group *Group,
	namePrefix string, numConnections int) (nodeID int) {
	name := namePrefix + fmt.Sprintf("%03d", len(group.ids))
	nodeID = createNode(nodeDb, name)
	for i := 0; i < numConnections && i < len(group.ids); i++ {
		addLink(nodeDb, nodeID, group.ids[i])
	}
	return nodeID
}
func TestGenerateGroupName(t *testing.T) {
	resetGroupNameGenerator()
	got := generateGroupName()
	want := "cliquetool00001"
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
	got = generateGroupName()
	want = "cliquetool00002"
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
	resetGroupNameGenerator()
	got = generateGroupName()
	want = "cliquetool00001"
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func TestCountSharedNodes(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	group1 := &groupDb.groups[0]
	group2 := &groupDb.groups[1]

	// No overlap between groups to start
	overlap := countSharedNodes(group1, group2)
	if overlap != 0 {
		t.Errorf("Expected overlap of 0, %d", overlap)
	}
	// Add 1 from group2 to group1
	addNodeToGroup(&nodeDb, group1, group2.ids[0], false)
	overlap = countSharedNodes(group1, group2)
	if overlap != 1 {
		t.Errorf("Expected overlap of 1, %d", overlap)
	}
	// Add 1 from group1 to group2
	addNodeToGroup(&nodeDb, group2, group1.ids[0], false)
	overlap = countSharedNodes(group1, group2)
	if overlap != 2 {
		t.Errorf("Expected overlap of 2, %d", overlap)
	}
}
func TestRenameGroups(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 4, "node")
	for i := 0; i < 4; i++ {
		if len(nodeDb.nodes[i].links) != 3 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 3)
		}
	}
	renameGroups(&groupDb, "testing")
	if groupDb.groups[0].groupName != "testing00000" {
		t.Errorf("Group has name '%s' instead of 'testing00000'", groupDb.groups[0].groupName)
	}
}

func TestSortGroups(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 4, "group1node")
	for i := 0; i < 4; i++ {
		if len(nodeDb.nodes[i].links) != 3 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 3)
		}
	}
	//dumpDb(nodeDb)
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	//dumpDb(nodeDb)
	for i := 4; i < 14; i++ {
		if len(nodeDb.nodes[i].links) != 9 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 9)
		}
	}
	//dumpGroups(&nodeDb, &groupDb)
	sortGroupDb(&groupDb)
	//dumpGroups(&nodeDb, &groupDb)
	//dumpDb(nodeDb)
	if len(groupDb.groups[0].ids) != 10 {
		t.Errorf("Expected group of size 10 to be first")
	}
	if len(groupDb.groups[1].ids) != 4 {
		t.Errorf("Expected group of size 4 to be second")
	}
}

func TestNodeCountsThatLinkToGroup(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 4, "group1node")
	for i := 0; i < 4; i++ {
		if len(nodeDb.nodes[i].links) != 3 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 3)
		}
	}
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	for i := 4; i < 14; i++ {
		if len(nodeDb.nodes[i].links) != 9 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 9)
		}
	}
	// Add some extra groups to the db just for noise
	for i := 11; i < 30; i++ {
		createConnectedNodes(&nodeDb, &groupDb, i, fmt.Sprintf("group%dnode", i))
	}
	sortGroupDb(&groupDb)
	//dumpGroups(&nodeDb, &groupDb)
	nGroups := len(groupDb.groups)
	if len(groupDb.groups[nGroups-2].ids) != 10 {
		t.Errorf("Expected group of size 10 to be first")
	}
	if len(groupDb.groups[nGroups-1].ids) != 4 {
		t.Errorf("Expected group of size 4 to be second")
	}
	// We now have two complete groups of size 10 and 4.
	group := &groupDb.groups[1]
	nodeCounts := nodeCountsThatLinkToGroup(&nodeDb, group, 1)
	if len(nodeCounts) != 0 {
		t.Errorf("Expected nodeCountsThatLinkToGroup to return empty array")
	}
	// Add a node that links to 3 of the 4.
	newNodeID := createNewNodeAndConnectToGroup(&nodeDb, &groupDb, group, "group1node_later1_", 2)
	nodeCounts = nodeCountsThatLinkToGroup(&nodeDb, group, 1)
	// Should be one entry for the just-added node
	if len(nodeCounts) != 1 {
		t.Errorf("Expected nodeCountsThatLinkToGroup to return 1")
	}
	if nodeCounts[0].cnt != 2 {
		t.Errorf("Expected nodeCounts[0].cnt to be 2")
	}
	if nodeCounts[0].id != newNodeID {
		t.Errorf("Expected id to be %d not %d", newNodeID, nodeCounts[0].id)
	}
	// Change min connected to 3
	nodeCounts = nodeCountsThatLinkToGroup(&nodeDb, group, 3)
	if len(nodeCounts) != 0 {
		t.Errorf("Expected nodeCountsThatLinkToGroup to return 0")
	}
	// Add fully connected
	createNewNodeAndConnectToGroup(&nodeDb, &groupDb, group, "group1node_later2_", 4)
	nodeCounts = nodeCountsThatLinkToGroup(&nodeDb, group, 1)
	if len(nodeCounts) != 2 {
		t.Errorf("Expected nodeCountsThatLinkToGroup to return array of size 2, not %d", len(nodeCounts))
	}
}

func TestNodesThatLinkToAllOfGroup(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 4, "group1node")
	for i := 0; i < 4; i++ {
		if len(nodeDb.nodes[i].links) != 3 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 3)
		}
	}
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	for i := 4; i < 14; i++ {
		if len(nodeDb.nodes[i].links) != 9 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 9)
		}
	}
	//dumpGroups(&nodeDb, &groupDb)
	sortGroupDb(&groupDb)
	//dumpGroups(&nodeDb, &groupDb)
	//dumpDb(nodeDb)
	if len(groupDb.groups[0].ids) != 10 {
		t.Errorf("Expected group of size 10 to be first")
	}
	if len(groupDb.groups[1].ids) != 4 {
		t.Errorf("Expected group of size 4 to be second")
	}
	// We now have two complete groups of size 10 and 4.
	// Add a node that links to 3 of the 4.
	group := &groupDb.groups[1]
	createNewNodeAndConnectToGroup(&nodeDb, &groupDb, group, "group1node_later1_", 3)
	// No new nodes since we only connected to 3 or 4
	nodes := nodesThatLinkToAllOfGroup(&nodeDb, group)
	if len(nodes) != 0 {
		t.Errorf("Expected nodesThatLinkToAllOfGroup to return empty array")
	}
	// Add fully connected
	createNewNodeAndConnectToGroup(&nodeDb, &groupDb, group, "group1node_later2_", 4)
	nodes = nodesThatLinkToAllOfGroup(&nodeDb, group)
	//oups(&nodeDb, &groupDb)
	//dumpDb(nodeDb)
	if len(nodes) != 1 {
		t.Errorf("Expected nodesThatLinkToAllOfGroup to return array of size 1, not %d", len(nodes))
	}

	// Now also test nodesThatLinkToGroup
	nodes = nodesThatLinkToGroup(&nodeDb, group, 3)
	// This should catch both of the new nodes we added.
	if len(nodes) != 2 {
		t.Errorf("Expected nodesThatLinkToAllOfGroup to return array of size 2, not %d", len(nodes))
	}
}

func TestUpdateGroupStats(t *testing.T) {
	var groupDb GroupDb

	// Create a group of 10 nodes with 9 connections each, 90 total links
	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	group := &groupDb.groups[0]
	updateGroupLinkCounts(&nodeDb, group)

	updateGroupStats(&nodeDb, group)
	if group.totalNodeLinks != 90 {
		t.Errorf("Expected totalNodeLinks of 90, not %d", group.totalNodeLinks)
	}
	if group.density < 0.999 {
		t.Errorf("Expected density of 1.0, not %v", group.density)
	}
}

func TestUpdateGroupLinkCounts(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	group := &groupDb.groups[0]
	updateGroupLinkCounts(&nodeDb, group)
	if len(group.ids) != 10 {
		t.Errorf("Expected group 10, not %d", len(group.ids))
	}
	for i := 0; i < 10; i++ {
		if len(nodeDb.nodes[i].links) != 9 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 9)
		}
	}
	if len(group.linkCounts) != 10 {
		t.Errorf("Gruop has %d linkCount instead of %d", len(group.linkCounts), 10)
	}

	// Remove node from start
	idToRemove := group.ids[0]
	removeNodeFromGroup(&nodeDb, group, idToRemove)
	if len(group.ids) != 9 {
		t.Errorf("Expected group size 9 after remove, not %d", len(group.ids))
	}
	if isInArray(group.ids, idToRemove) {
		t.Errorf("Id %d not removed", idToRemove)
	}
	updateGroupLinkCounts(&nodeDb, group)
	if len(group.linkCounts) != 9 {
		t.Errorf("Gruop has %d linkCount instead of %d", len(group.linkCounts), 9)
	}
}

func TestRemoveNodeFromGroup(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}
	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	group := &groupDb.groups[0]
	if len(group.ids) != 10 {
		t.Errorf("Expected group 10, not %d", len(group.ids))
	}
	for i := 0; i < 10; i++ {
		if len(nodeDb.nodes[i].links) != 9 {
			t.Errorf("Node %d has %d links instead of %d", i, len(nodeDb.nodes[i].links), 9)
		}
	}

	// Remove from start
	idToRemove := group.ids[0]
	removeNodeFromGroup(&nodeDb, group, idToRemove)
	if len(group.ids) != 9 {
		t.Errorf("Expected group size 9 after remove, not %d", len(group.ids))
	}
	if isInArray(group.ids, idToRemove) {
		t.Errorf("Id %d not removed", idToRemove)
	}
	// Remove from middle
	idToRemove = group.ids[4]
	removeNodeFromGroup(&nodeDb, group, idToRemove)
	if len(group.ids) != 8 {
		t.Errorf("Expected group size 8 after remove, not %d", len(group.ids))
	}
	if isInArray(group.ids, idToRemove) {
		t.Errorf("Id %d not removed", idToRemove)
	}
	// Remove from end
	idToRemove = group.ids[len(group.ids)-1]
	removeNodeFromGroup(&nodeDb, group, idToRemove)
	if len(group.ids) != 7 {
		t.Errorf("Expected group size 7 after remove, not %d", len(group.ids))
	}
	if isInArray(group.ids, idToRemove) {
		t.Errorf("Id %d not removed", idToRemove)
	}
}

func TestCountGroupOverlap(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	createConnectedNodes(&nodeDb, &groupDb, 10, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 10, "group2node")
	group1 := &groupDb.groups[0]
	group2 := &groupDb.groups[1]

	// No overlap between groups to start
	overlap := countGroupOverlap(group1, group2)
	if overlap != 0 {
		t.Errorf("Expected overlap of 0, %d", overlap)
	}
	// Add 1 from group2 to group1
	addNodeToGroup(&nodeDb, group1, group2.ids[0], false)
	overlap = countGroupOverlap(group1, group2)
	if overlap != 1 {
		t.Errorf("Expected overlap of 1, %d", overlap)
	}
	// Add 1 from group1 to group2
	addNodeToGroup(&nodeDb, group2, group1.ids[0], false)
	overlap = countGroupOverlap(group1, group2)
	if overlap != 2 {
		t.Errorf("Expected overlap of 2, %d", overlap)
	}
}

func TestIsGroupSubset(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	createConnectedNodes(&nodeDb, &groupDb, 5, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 5, "group2node")
	group1 := &groupDb.groups[0]
	group2 := &groupDb.groups[1]

	// No overlap between groups to start
	if isGroupSubset(group1, group2) || isGroupSubset(group2, group1) {
		t.Errorf("Groups are not subsets")
	}
	// Add almost all nodes of group2 to group1
	for i := 0; i < 4; i++ {
		addNodeToGroup(&nodeDb, group1, group2.ids[i], false)
		if isGroupSubset(group1, group2) || isGroupSubset(group2, group1) {
			t.Errorf("Groups are not subsets")
		}
	}
	// Add final node of group2 to group1 to make group2 a subset of group1
	addNodeToGroup(&nodeDb, group1, group2.ids[len(group2.ids)-1], false)
	if !isGroupSubset(group1, group2) {
		fmt.Printf("Group2: %v\nGroup1: %v\n", group2.ids, group1.ids)
		t.Errorf("Group2 should be subset of Group1")
	}
	if isGroupSubset(group2, group1) {
		fmt.Printf("Group2: %v\nGroup1: %v\n", group2.ids, group1.ids)
		t.Errorf("Group1 should not be subset of Group2")
	}
}

func TestRemoveGroupSubsets(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	createConnectedNodes(&nodeDb, &groupDb, 5, "group1node")
	createConnectedNodes(&nodeDb, &groupDb, 5, "group2node")

	// No overlap between groups to start
	if isGroupSubset(&groupDb.groups[0], &groupDb.groups[1]) || isGroupSubset(&groupDb.groups[1], &groupDb.groups[0]) {
		t.Errorf("Groups are not subsets")
	}
	if len(groupDb.groups) != 2 {
		t.Errorf("Should be 2 groups")
	}
	removeGroupSubsets(&groupDb, 1)
	if len(groupDb.groups) != 2 {
		t.Errorf("Should be 2 groups")
	}
	// Add  all nodes of group2 to group1
	for i := 0; i < 5; i++ {
		addNodeToGroup(&nodeDb, &groupDb.groups[0], groupDb.groups[1].ids[i], false)
	}
	if !isGroupSubset(&groupDb.groups[0], &groupDb.groups[1]) {
		t.Errorf("Group2 should be subset of Group1")
	}
	nSubsets := removeGroupSubsets(&groupDb, 1)
	if nSubsets != 1 {
		t.Errorf("Should have returned 1 from removeGroupSubsets")
	}
	if len(groupDb.groups) != 1 {
		fmt.Printf("\nGroup2: %v\nGroup1: %v\n", groupDb.groups[1].ids, groupDb.groups[0].ids)
		t.Errorf("Should be 1 group after removeGroupSubsets")
	}
}
