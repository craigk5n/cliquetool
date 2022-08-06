package main

// Group utilities

import (
	"fmt"
	"sort"
)

// NodeCount defines a structure used to sort nodes within a group.
type NodeCount struct {
	id       int
	cnt      int
	totalCnt int
}

var groupIDCounter int32 = 0

// only used for debugging...
func nodeIdsToString(db *NodeDb, ids []int) string {
	ret := "[" + db.nodes[ids[0]].externalName
	for i := 1; i < len(ids); i++ {
		ret = ret + fmt.Sprintf(",%s", db.nodes[ids[i]].externalName)
	}
	ret = ret + "]"
	return ret
}

func generateGroupID() int32 {
	groupIDCounter++
	return groupIDCounter
}

func resetGroupNameGenerator() {
	groupIDCounter = 0
}

func generateGroupName() string {
	return fmt.Sprintf("cliquetool%05d", generateGroupID())
}

// Rename the groups using the specified prefix.
func renameGroups(groupDb *GroupDb, prefix string) {
	for i := 0; i < len(groupDb.groups); i++ {
		groupDb.groups[i].groupName = fmt.Sprintf("%s%05d", prefix, i)
		//fmt.Printf("new group #%d name: %s\n", i, groupDb.groups[i].groupName)
	}
}

// Sort the groups in a GroupDb, placing the groups with the largest
// number of nodes at the beginning.
func sortGroupDb(groupDb *GroupDb) {
	sort.Slice(groupDb.groups[:], func(i, j int) bool {
		return len(groupDb.groups[i].ids) > len(groupDb.groups[j].ids)
	})
}

// Increment the count for links to the specified node id
// for this NodeCount structure.
func addLinkToNodeCount(db *NodeDb, nodeCount []NodeCount, id int) []NodeCount {
	//var ret []NodeCount = nodeCount
	found := false

	/* search for node in existing array */
	for j := 0; j < len(nodeCount) && !found; j++ {
		if nodeCount[j].id == id {
			/* found it */
			nodeCount[j].cnt++
			found = true
		}
	}
	if !found {
		/* append to end of list */
		var newNodeCount = NodeCount{id: id, cnt: 1, totalCnt: len(db.nodes[id].links)}
		nodeCount = append(nodeCount, newNodeCount)
	}
	return nodeCount
}

// Create a list of node ids of all the nodes that link to every member
// of the specified group.
func nodesThatLinkToAllOfGroup(db *NodeDb, group *Group) []int {
	return nodesThatLinkToGroup(db, group, len(group.ids))
}

// Create a list of node ids of all the nodes that link to members
// of the specified group.  Only include nodes in the result that
// link to at least minNumLinksToGroup nodes.
func nodesThatLinkToGroup(db *NodeDb, group *Group, minNumLinksToGroup int) []int {
	var nodeCount []NodeCount
	ret := []int{}

	nodeCount = nodeCountsThatLinkToGroup(db, group, minNumLinksToGroup)

	for i := 0; i < len(nodeCount); i++ {
		if nodeCount[i].cnt >= minNumLinksToGroup {
			ret = append(ret, nodeCount[i].id)
		} else {
			break
		}
	}

	return ret
}

// Create an array of NodeCount structures that lists all nodes that link
// to members of the specified group.  Only include nodes in the result that
// link to at least minNumLinksToGroup nodes.
// Caller should free returned result.
func nodeCountsThatLinkToGroup(db *NodeDb, group *Group, minNumLinksToGroup int) []NodeCount {
	var nodeCount []NodeCount = []NodeCount{}

	/*
	 * Use the links of current group members to find candidates for
	 * other group members.  Scan through all links that the
	 * current group nodes have and count how many times they
	 * reference each node.
	 */
	for i := 0; i < len(group.ids); i++ {
		/* get all links for node id ids[i] */
		nodeId := group.ids[i]
		if nodeId >= len(db.nodes) {
			fmt.Printf("Programmer Bug#457123\n")
			continue
		}
		for j := 0; j < len(db.nodes[nodeId].links); j++ {
			/* Make sure this link does not reference a current group node */
			if !isInArray(group.ids, db.nodes[nodeId].links[j]) {
				/* TODO: only add if this new node has minSize links */
				nodeCount = addLinkToNodeCount(db, nodeCount, db.nodes[nodeId].links[j])
			}
		}
	}
	/*
	 * If nNodes is 0, then this group has no links to anyone outside
	 * if the group.  If nNodes is non-zero, then check the links counts
	 * of each to see if any have links to each current group node.
	 */
	if len(nodeCount) > 0 {
		/* sort the nodes by number of links to it */
		sort.Slice(nodeCount[:], func(i, j int) bool {
			return cmpNodeCountLinkCounts(nodeCount[i], nodeCount[j])
		})
		// Scan until we find the last that links to the specified limit.
		// We will then truncate the end of the list that did not meet threshold.
		nret := 0
		for i := 0; i < len(nodeCount); i++ {
			if nodeCount[i].cnt >= minNumLinksToGroup {
				nret = i + 1
			} else {
				break
			}
		}
		// truncate
		if nret < len(nodeCount) {
			nodeCount = nodeCount[:nret]
		}
	}

	return nodeCount
}

// sort from highest to lowest
func cmpNodeCountLinkCounts(a NodeCount, b NodeCount) bool {
	if a.cnt < b.cnt {
		return false
	} else if a.cnt > b.cnt {
		return true
	} else {
		/* secondary sort is the internal to external ratio */
		/* or simply the lower number of total links (since a and b */
		/* have the same number of group links) */
		if a.totalCnt < b.totalCnt {
			return true
		} else if a.totalCnt > b.totalCnt {
			return false
		}
		return false
	}
}

// Calcuate the percent overlap between two groups.  This value
// is between 0.0 and 1.0 and is the number of common nodes divided
// by the number of nodes in the smaller group.
func percentGroupOverlap(group1 *Group, group2 *Group) float64 {
	smallSize := len(group1.ids)
	if len(group2.ids) > smallSize {
		smallSize = len(group2.ids)
	}
	overlap := countGroupOverlap(group1, group2)
	ret := float64(overlap) / float64(smallSize)
	return ret
}

// Count the number of common nodes between two groups.
func countGroupOverlap(group1 *Group, group2 *Group) int {
	overlap := 0

	for i := 0; i < len(group1.ids); i++ {
		if isInArray(group2.ids, group1.ids[i]) {
			overlap++
		}
	}

	return overlap
}

// Count how many nodes are present in both the provided groups.
func countSharedNodes(group *Group, group2 *Group) int {
	ret := 0

	for i := 0; i < len(group2.ids); i++ {
		if isInArray(group.ids, group2.ids[i]) {
			ret++
		}
	}
	return ret
}

// Count the number of links that the specified node ID has to the specified group.
func countLinksFromGroupToNode(nodeDb *NodeDb, group *Group, id int) int {
	cnt := 0

	for i := 0; i < len(group.ids); i++ {
		/* skip ourself, of course */
		if group.ids[i] == id {
			continue
		} else if nodeLinksTo(*nodeDb, group.ids[i], id) {
			cnt++
		}
	}
	return cnt
}

// Remove all subgroups in the specified groupdb. Returns the
// number of subsets removed.
func removeGroupSubsets(groupDb *GroupDb, verbose int) int {
	if verbose > 0 {
		logMessage("Removing subsets...")
	}

	subsets := make([]bool, len(groupDb.groups))
	for i := 0; i < len(groupDb.groups); i++ {
		group := groupDb.groups[i]
		/* find all subsets of this group */
		for j := 0; j < len(groupDb.groups); j++ {
			if i != j && isGroupSubset(&group, &groupDb.groups[j]) {
				logMessage(fmt.Sprintf("Group %s is a subset of %s",
					groupDb.groups[j].groupName, group.groupName))
				/* mark smaller as subset */
				subsets[j] = true
			}
		}
	}

	/* count up number of subsets found */
	nSubsets := 0
	for i := 0; i < len(groupDb.groups); i++ {
		if subsets[i] {
			nSubsets++
		}
	}

	newNum := len(groupDb.groups) - nSubsets
	newGroups := []Group{}
	for i := 0; i < len(groupDb.groups); i++ {
		if !subsets[i] {
			newGroups = append(newGroups, groupDb.groups[i])
		}
	}

	logMessage(fmt.Sprintf("Done removing subsets: %d -> %d groups",
		len(groupDb.groups), newNum))
	if verbose > 0 {
		logMessage(fmt.Sprintf("Done removing subsets: %d -> %d groups",
			len(groupDb.groups), newNum))
	}
	groupDb.groups = newGroups
	return nSubsets
}

// Is group2 a subset of group1?
// If so, then group1 contains every member of group2.
func isGroupSubset(group1 *Group, group2 *Group) bool {
	/*
	* If group1 has less members than group2, then it can't possibly
	* contain all members.
	 */
	if len(group1.ids) < len(group2.ids) {
		return false
	}

	for i1 := 0; i1 < len(group2.ids); i1++ {
		/* does group1 contain this group2 node? if not, then not a subset */
		if !isInArray(group1.ids, group2.ids[i1]) {
			return false
		}
	}

	/*
	* Group1 contained every member of group2.  Therefore, group2 is
	* a subset of group1.
	 */
	return true
}

// Has the specified group (or a duplicate of it) already been
// add to the GroupDb?  This keeps us from adding subgroups of existing
// groups to our list of groups.
//
// Performance notes:
// The largest chunk of time for finding dense groups is spent in
// group subset searching (isInArray, etc.)
func groupIsSubsetOfExisting(nodeDb *NodeDb, groupDb *GroupDb, group *Group, mustBeIdentical bool) bool {
	for i := len(groupDb.groups) - 1; i >= 0; i-- {
		if !mustBeIdentical || len(groupDb.groups[i].ids) == len(group.ids) {
			if isGroupSubset(&groupDb.groups[i], group) {
				return true
			}
		}
	}
	/* no identical or larger group that this is a subset of found */
	return false
}

// Add a group to the Group Db.
// This will copy the data from the group parameter and will
// make a copy of the array pointers as well (ids, links).
func addGroup(nodeDb *NodeDb, groupDb *GroupDb, group *Group, checkForSubset bool) bool {
	/* If this group exists, or this is a subgroup of an existing group, */
	/* then do not add this new group. */
	if checkForSubset && groupIsSubsetOfExisting(nodeDb, groupDb, group, false) {
		return false /*  not added */
	}
	groupDb.mutex.Lock()
	defer groupDb.mutex.Unlock()
	groupDb.groups = append(groupDb.groups, *group)
	return true
}

// Make a copy of the specified group and then add the new node
// to it to make the new group.
func cloneAndAddToGroup(nodeDb *NodeDb, group *Group, id int) (*Group, error) {
	var ret Group

	group.mutex.Lock()
	defer group.mutex.Unlock()
	/* don't allow adding duplicate ids to a group */
	if isInArray(group.ids, id) {
		return nil, fmt.Errorf("cannot add duplicate ID")
	}

	ret.groupName = generateGroupName()

	ret.ids = make([]int, len(group.ids)+1)
	copy(ret.ids, group.ids)
	ret.ids[len(group.ids)] = id
	ret.statsAreCurrent = false

	/* Note: linkCounts is now out-of-date, but we will update when we */
	/* are done finding all the dense groups */

	return &ret, nil
}

// Make a copy of the specified group except exclude the specified
// node id from the new copy.
func cloneAndSubtractFromGroup(nodeDb *NodeDb, group *Group, id int) *Group {
	var ret Group
	ret.groupName = generateGroupName()

	group.mutex.Lock()
	defer group.mutex.Unlock()
	ret.ids = make([]int, len(group.ids)-1)
	ret.linkCounts = make([]int, len(group.ids)-1)

	j := 0
	for i := 0; i < len(group.ids); i++ {
		if group.ids[i] == id {
			/* ignore this node */
		} else if j < len(ret.ids) {
			if j < len(group.linkCounts)-1 {
				ret.linkCounts[j] = group.linkCounts[i]
			}
			ret.ids[j] = group.ids[i]
			j++
		} else {
			/* invalid id passed in */
			fmt.Printf("cloneAndSubtractFromGroup: id=%d not in group %s\n", id, group.groupName)
		}
	}

	if len(ret.ids) != j {
		fmt.Printf("Programmer error #923896\n")
	}
	// Note: linkCounts is now out-of-date, but we will update when we *
	// are done finding all the dense groups

	return &ret
}

/* no need to sort nodes since it already was sorted */

// Make a deep copy of the specified GroupDb.
func cloneGroupDb(groupDb *GroupDb) *GroupDb {
	var ret GroupDb
	ret.groups = make([]Group, len(groupDb.groups))

	groupDb.mutex.Lock()
	defer groupDb.mutex.Unlock()

	for i := 0; i < len(groupDb.groups); i++ {
		groupDb.groups[i].mutex.Lock()
		ret.groups[i] = Group{
			groupName:       groupDb.groups[i].groupName,
			totalNodeLinks:  groupDb.groups[i].totalNodeLinks,
			density:         groupDb.groups[i].density,
			statsAreCurrent: groupDb.groups[i].statsAreCurrent,
			status:          groupDb.groups[i].status,
		}
		ret.groups[i].ids = make([]int, len(groupDb.groups[i].ids))
		ret.groups[i].linkCounts = make([]int, len(groupDb.groups[i].linkCounts))
		copy(ret.groups[i].ids, groupDb.groups[i].ids)
		copy(ret.groups[i].linkCounts, groupDb.groups[i].linkCounts)
		ret.groups[i].statsAreCurrent = false
		groupDb.groups[i].mutex.Unlock()
	}
	return &ret
}

// Add a node to a group if it is not already in the group.
// The stats (linkCounts, etc.) will be out of date, and statsAreCurrent will
// be set to false.  Invoke updateGroupStats later to update these fields.
func addNodeToGroup(nodeDb *NodeDb, group *Group, id int, doSortNodes bool) {
	/* first, make sure id is not already in group */
	if isInArray(group.ids, id) {
		/* This id is already in group */
		return
	}

	group.mutex.Lock()
	group.ids = append(group.ids, id)
	group.linkCounts = append(group.linkCounts, 0)
	group.statsAreCurrent = false
	group.mutex.Unlock()

	/* now sort nodes in group */
	if doSortNodes {
		sort.Ints(group.ids[:])
	}
	/* Note: linkCounts is now out-of-date, but we will update when we */
	/* are done finding all the dense groups */
}

// Update the statistics in the Group structure (if they are out of date).
func updateGroupStats(db *NodeDb, group *Group) {
	// only do this if they are out-of-date
	if group.statsAreCurrent {
		return
	}

	nodeLinks := 0
	totalLinks := 0

	for i := 0; i < len(group.ids); i++ {
		nodeId := group.ids[i]
		nodeLinks += countLinksFromGroupToNode(db, group, nodeId)
		totalLinks += len(db.nodes[nodeId].links)
	}

	group.groupLinks = nodeLinks
	group.totalNodeLinks = totalLinks

	group.density = float64(group.groupLinks) /
		(float64(len(group.ids)) * float64(len(group.ids)-1))

	group.statsAreCurrent = true
}

// Update the linkCounts array in the group.
// Call this after adding or removing a node to a group.
func updateGroupLinkCounts(nodeDb *NodeDb, group *Group) {
	if len(group.linkCounts) != len(group.ids) {
		group.linkCounts = make([]int, len(group.ids))
	}
	for i := 0; i < len(group.ids); i++ {
		group.linkCounts[i] = countLinksFromGroupToNode(nodeDb,
			group, group.ids[i])
	}
}

// Remove a node from a group.
func removeNodeFromGroup(nodeDb *NodeDb, group *Group, id int) {
	ind := -1
	for i := 0; i < len(group.ids) && ind < 0; i++ {
		if group.ids[i] == id {
			ind = i
		}
	}

	if ind < 0 {
		return // not found
	}

	newIds := make([]int, len(group.ids)-1)
	//newLinksCounts := make([]int, len(group.ids)-1)
	j := 0
	for i := 0; i < len(group.ids); i++ {
		if i != ind {
			newIds[j] = group.ids[i]
			//newLinksCounts[j] = group.linkCounts[i]
			j++
		}
	}
	group.ids = newIds
	//group.linkCounts = newLinksCounts
	group.statsAreCurrent = false
	//fmt.Printf("Group Ids after remove: %v\n", group.ids)
}

// Count how many groups each node belongs to and store the info
// in the NodeDb.  This should be called after all groups have been
// loaded/calculated before running any verification/report.
func updateNodeGroupCounts(db *NodeDb, groupDb *GroupDb) {
	/* First, reset all the counts back to 0. */
	for i := 0; i < len(db.nodes); i++ {
		db.nodes[i].numGroups = 0
	}
	// Now, loop through each group
	for i := 0; i < len(groupDb.groups); i++ {
		// Loop through each member of the group and bump their group count
		for j := 0; j < len(groupDb.groups[i].ids); j++ {
			id := groupDb.groups[i].ids[j]
			db.nodes[id].numGroups++
		}
	}
}

// Debugging function
func groupHasDuplicates(group *Group) bool {
	for _, id := range group.ids {
		cnt := 0
		for _, id2 := range group.ids {
			if id == id2 {
				cnt++
			}
		}
		if cnt > 1 {
			fmt.Printf("Found duplicate ID %d in group %s\n", id, group.groupName)
			return true
		}
	}
	return false
}

// Debugging function
func checkGroupsForDups(groupDb *GroupDb) {
	for i := 0; i < len(groupDb.groups); i++ {
		groupHasDuplicates(&groupDb.groups[i])
	}
}
