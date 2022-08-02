package main

import (
	"testing"
)

func initTests() *NodeDb {
	nodeDb := initNodeDb()
	return &nodeDb
}

func TestInitNodeDb(t *testing.T) {
	nodeDb := initTests()
	got := len(nodeDb.nodes)
	want := 0
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func TestGetNode(t *testing.T) {
	nodeDb := initTests()
	name0 := "Node 0"
	name1 := "Node 1"
	// Add node
	id0 := createNode(nodeDb, name0)
	if id0 < 0 {
		t.Errorf("Failed to create %s", name0)
	}
	if nodeDb.nodes[0].externalName != name0 {
		t.Errorf("NodeDb node 0 external name wrong")
	}
	// Add 2nd node
	id1 := createNode(nodeDb, name1)
	if id1 != 1 {
		t.Errorf("Failed to create %s", name1)
	}
	if nodeDb.nodes[1].externalName != name1 {
		t.Errorf("NodeDb node 1 external name wrong")
	}
	// Add link between node0 and node1
	//fmt.Println("Calling addLink")
	err := addLink(nodeDb, id0, id1)
	if err != nil {
		t.Errorf("Error adding link: %s", err)
	}
	//fmt.Println("addLink done")
	// Verify we have a link on both nodes.
	if !nodeLinksTo(*nodeDb, id0, id1) {
		t.Errorf("Missing link from %s to %s", name0, name1)
	}
	if len(nodeDb.nodes[0].links) != 1 {
		t.Errorf("Missing link from Node 0 to Node 1")
	}
	if len(nodeDb.nodes[1].links) != 1 {
		t.Errorf("Missing link from Node 1 to Node 0")
	}
	// Try adding original node again. Should get existing node ID back.
	id0B := createNode(nodeDb, name0)
	if id0B < 0 {
		t.Errorf("Failed to create %s", name0)
	}
	if id0B != id0 {
		t.Errorf("Created new node instead of returning existing node")
	}
}
