package main

// Functions for handling nodes (also called vertices in graph theory).
// Note that we only consider an undirected graph.

import (
	"errors"
	"fmt"
)

// Node defines a structure to hold a node including the external
// name used to identify the node and what other nodes a node
// is connected to.  The node ID is the position (slide index)
// in the NodeDb.nodes slice.
type Node struct {
	// Internal Id of node.  This is generated internally and has no
	// relation to the original external Id found in the data file.
	id int
	// The external name or Id (may be an integer id)
	externalName string
	// Slice of Ids of who this Node links to
	links []int
	// Number of groups this node is a member of
	numGroups int
}

// NodeDb defines a structure to be the Node database.
type NodeDb struct {
	// Slice of nodes.  The node.id is the slice index location.
	nodes []Node
	// Hashtable to allow external name lookups
	hash map[string]int
}

// Dump the contents of the NodeDb to stdout. (Used for debugging.)
func dumpDb(nodeDb NodeDb) {
	fmt.Printf("** NodeDb **\n")
	for i := 0; i < len(nodeDb.nodes); i++ {
		var n Node = nodeDb.nodes[i]
		fmt.Printf("Node#%d: %p id=%d, name='%s', nlinks=%d, links=", i, &nodeDb.nodes[i], n.id, n.externalName, len(n.links))
		if len(nodeDb.nodes[i].links) == 0 {
			fmt.Print("none")
		}
		for j := 0; j < len(nodeDb.nodes[i].links); j++ {
			if j > 0 {
				fmt.Print(",")
			}
			fmt.Print(nodeDb.nodes[i].links[j])
		}
		fmt.Println()
	}
}

// Dump the contents of the Node to stdout. (Used for debugging.)
func printNode(n *Node) {
	fmt.Printf("Node: addr=%p  id=%d  nLinks=%d\n", n, n.id, len(n.links))
}

func initNodeDb() NodeDb {
	var nodeDb NodeDb
	// init empty slice of Node
	nodeDb.nodes = []Node{}
	nodeDb.hash = make(map[string]int)

	return nodeDb
}

// Get a Node pointer, or return nil if one does not exist
// for this id.
/* OBE
func getNode(nodeDb NodeDb, id int) (*Node, error) {
	if id >= len(nodeDb.nodes) {
		return nil, errors.New("No such node")
	}

	return &(nodeDb.nodes[id]), nil
}
*/

// Create a new Node with the given name.  Return the ID.
func createNode(nodeDb *NodeDb, name string) int {
	debugMessage(fmt.Sprintf("In createNode with name=%s", name))
	var ret Node
	lookupID, e := getNodeIDForExternalName(*nodeDb, name)
	if e == nil && lookupID >= 0 {
		debugMessage(fmt.Sprintf("  Returning existing id=%d\n", lookupID))
		return lookupID
	}
	if lookupID < 0 {
		debugMessage(fmt.Sprintf("  Not found name=%s, creating now", name))
		lookupID = len(nodeDb.nodes)
	}
	debugMessage(fmt.Sprintf("  Allocating new Node with id=%d for name=%s", lookupID, name))
	ret.id = lookupID
	ret.externalName = name
	// start with empty slice of int Ids
	ret.links = []int{}
	ret.externalName = name
	ret.numGroups = 0
	// append to slize of nodes
	nodeDb.nodes = append(nodeDb.nodes, ret)
	// Now add to hash table using external name as key
	nodeDb.hash[ret.externalName] = ret.id
	debugMessage(fmt.Sprintf("  Num nodes is now %d", len(nodeDb.nodes)))
	return lookupID
}

// Add a link to a node.
func addLinkToNode(nodeDb *NodeDb, sourceNodeID int, linkedNodeID int) error {
	debugMessage(fmt.Sprintf("Adding link to node width id=%d", sourceNodeID))
	maxValidID := len(nodeDb.nodes) - 1
	// Verify both are valid IDs
	if sourceNodeID < 0 || sourceNodeID > maxValidID {
		debugMessage("Invalid sourceNodeID")
		return fmt.Errorf("invalid sourceNodeID %d", sourceNodeID)
	}
	if linkedNodeID < 0 || linkedNodeID > maxValidID {
		debugMessage("Invalid linkedNodeID")
		return fmt.Errorf("invalid linkedNodeID %d", linkedNodeID)
	}
	// Make sure we don't already have this link
	debugMessage(fmt.Sprintf("  Existing links for node: %d", len(nodeDb.nodes[sourceNodeID].links)))
	for i := 0; i < len(nodeDb.nodes[sourceNodeID].links); i++ {
		if nodeDb.nodes[sourceNodeID].links[i] == linkedNodeID {
			// Found link already
			debugMessage("  (Link already exists)")
			return nil
		}
	}

	// not found -> add to end of list
	nodeDb.nodes[sourceNodeID].links = append(nodeDb.nodes[sourceNodeID].links, linkedNodeID)
	debugMessage(fmt.Sprintf("  New links for node id=%d: %d", sourceNodeID, len(nodeDb.nodes[sourceNodeID].links)))
	return nil
}

// Add a link between two existing nodes.  The nodes may or may not already exist.
func addLink(nodeDb *NodeDb, id1 int, id2 int) error {
	// Verify both are valid IDs
	maxValidID := len(nodeDb.nodes) - 1
	if id1 < 0 || id1 > maxValidID {
		return fmt.Errorf("invalid sourceNodeId %d", id1)
	}
	if id2 < 0 || id2 > maxValidID {
		return fmt.Errorf("invalid linkedNodeId %d", id2)
	}
	//name1 := nodeDb.nodes[id1].externalName
	//name2 := nodeDb.nodes[id2].externalName
	err1 := addLinkToNode(nodeDb, id1, id2)
	err2 := addLinkToNode(nodeDb, id2, id1)
	if err1 == nil && err2 == nil {
		return nil
	} else if err1 != nil {
		return err1
	} else {
		return err2
	}
}

// Retrieve a Node id from the NodeDb by its external name.
func getNodeIDForExternalName(nodeDb NodeDb, name string) (int, error) {
	// Look it up in our map
	value, ok := nodeDb.hash[name]
	if ok {
		// Found.  Return the internal Id.
		return value, nil
	}
	// Not found
	return -1, errors.New("no such external Id")

}

// Determine if the two nodes are currently linked
func nodeLinksTo(nodeDb NodeDb, id1 int, id2 int) bool {
	n1 := nodeDb.nodes[id1]
	for i := 0; i < len(n1.links); i++ {
		if n1.links[i] == id2 {
			return true
		}
	}
	return false
}

// Create an abbreviated version of the external name.
// This is useful for fixed length columns in a plain/text report.
func nodeNameAbbr(node Node, length int) string {
	if len(node.externalName) <= length {
		return node.externalName
	}
	return node.externalName[0:length]
}
