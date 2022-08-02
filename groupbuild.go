package main

import (
	"fmt"
	"math"
	"sort"
	"time"
)

const MaxGroupOverlapThreshold float64 = 0.50 // 50%

// Take a set of groups and attempt to expand them by adding
// in nodes that meet the requirements specified in the parameter table.
// Normally, the input groups would be unmerged dense groups, but
// you could also pass in partially merged groups created from another
// application.
//
// The buildFast parameter (0/1) specifies how we build:
// 1 (fast): add the node with the most links to the group and move on
//   to add more nodes
// 2 (not fast): try adding each node that meets criteria in parameter
//   table and move on.  Compare results from the various paths and
//   pick the largest table.
func buildGroups(db *NodeDb, origGroups *GroupDb, paramTable GroupParameters, buildFast bool, minSize int, verbose int) *GroupDb {
	var ret *GroupDb
	var percent float64
	var perc = 0
	var perc2 int
	var now time.Time

	startTime := time.Now()
	lastUpdate := time.Now()

	logMessage(fmt.Sprintf("Building groups (fast=%t)...", buildFast))
	if verbose > 0 {
		fmt.Printf("Building groups (fast=%t)...\n", buildFast)
	}

	var groupDb GroupDb
	ret = &groupDb

	/*
	 * Make sure groups with highest node counts are first
	 * Note: this does affect the input groups, but not enough
	 * that we should create a whole new copy.
	 */
	sortGroupDb(origGroups)

	for i := 0; i < len(origGroups.groups); i++ {
		now = time.Now()
		percent = 100.0 * float64(i) / float64(len(origGroups.groups))
		perc2 = int(math.Floor(percent))
		if (perc2 != perc || now.UnixMilli() > (lastUpdate.UnixMilli()+5)) && verbose > 0 {
			perc = perc2
			if now != startTime && now.UnixMilli() > (lastUpdate.UnixMilli()+5000) && verbose > 0 {
				fmt.Printf("%d%% (%d of %d)", perc, i, len(origGroups.groups))
				secsSoFar := float64(now.UnixMilli()) - float64(startTime.UnixMilli())
				totalApprox := (100.0 / percent) * secsSoFar
				// 1st half is faster... so give more time
				totalApprox += 1.5 * (totalApprox * (1.0 - (percent / 100.0)))
				timeLeft := totalApprox - secsSoFar
				eta := int(timeLeft)
				etaH := eta / 3600
				etaM := (eta / 60) % 60
				etaS := eta % 60
				if verbose > 0 {
					fmt.Printf(" %02d:%02d:%02d remaining\n", etaH, etaM, etaS)
				}
				lastUpdate = now
			}
		}
		attemptToBuildGroup(db, ret, &origGroups.groups[i], &paramTable, buildFast, minSize)
	}

	// We may have created subsets if the subset was added prior to
	// its superset.
	removeGroupSubsets(ret, verbose)

	// Sort from largest to smallest
	sortGroupDb(ret)

	if verbose > 0 {
		fmt.Printf("%d%% (%d of %d)\n", 100, len(origGroups.groups),
			len(origGroups.groups))
		fmt.Printf("Done building groups.\n")
		now = time.Now()
		timeToComplete := int(now.UnixMilli() - startTime.UnixMilli())
		durH := timeToComplete / 3600
		durM := (timeToComplete / 60) % 60
		durS := timeToComplete % 60
		fmt.Printf("Build processing time: %02d:%02d:%02d\n", durH, durM, durS)
	}
	logMessage("Build processing complete")

	return ret
}

// Take the specified group and see if it can be added to.
func attemptToBuildGroup(db *NodeDb, groupDb *GroupDb, group *Group, paramTable *GroupParameters, buildFast bool, minSize int) {
	// If this group is a subset of an existing group, we can ignore it.
	if groupIsSubsetOfExisting(db, groupDb, group, false) {
		return
	}

	// Create a temporary group db.  We will store all new groups derived
	// from the original group in this temp group db.  After we have created
	// them all, then we will find the "most valuable" group and add
	// that one group.  This will be the largest group.  We then examine
	// the other groups.  If other groups don't have at least 50% overlap
	// with the "most valueable" group, they will also be added.
	var tempDb GroupDb

	if !expandGroup(db, &tempDb, group, paramTable, buildFast, minSize) {
		// We could not add to this group.  So, add it as is.
		addGroup(db, groupDb, group, true)
	} else {
		// We expanded the original group into one or more new groups.
		// Look for the largest.  Sort by size.
		sortGroupDb(&tempDb)
		// Add largest without modifying the original group name. */
		addGroup(db, groupDb, &tempDb.groups[0], true)
		if !buildFast {
			/*
			* Examine the rest to see if they  are unique enough to
			* also be added.  Only do this if we are not in buildFast mode.
			* This makes the process painfully slow on large datasets.
			 */
			for i := 1; i < len(tempDb.groups); i++ {
				hasOverlap := false
				for j := 0; j < len(tempDb.groups); j++ {
					if i != j && percentGroupOverlap(&tempDb.groups[i], &tempDb.groups[j]) >= MaxGroupOverlapThreshold {
						hasOverlap = true
					}
				}
				if !hasOverlap {
					/* Add, but modify group name */
					tempDb.groups[i].groupName = fmt.Sprintf("%s-%d", tempDb.groups[i].groupName, i)
				} else {
					tempDb.groups[i].groupName = fmt.Sprintf("%s%04d", "cliquetool", i)
				}
				addGroup(db, groupDb, &tempDb.groups[i], true)
			}
		}
	}
}

// Try to expand the specified group.
// We will do a link summary of this group to create an array that is sorted
// with nodes that links the most often at the start of the array.
// We will then add the node with the most links to this group.
func expandGroup(nodeDb *NodeDb, groupDb *GroupDb, group *Group, paramTable *GroupParameters, buildFast bool, minSize int) bool {
	ret := false

	minLinks := minLinksForGroupSize(paramTable, len(group.ids)+1)
	logMessage(fmt.Sprintf("Building group %s (size=%d)", group.groupName, len(group.ids)))

	potentials := nodesThatLinkToGroup(nodeDb, group, minLinks)
	logMessage(fmt.Sprintf("  found %d nodes with %d or more links to group: %v", len(potentials), minLinks, nodeIdsToString(nodeDb, group.ids)))
	if len(potentials) > 0 {
		/* sort front of list */
		potentials = sortPotentials(nodeDb, group, potentials)
		for i := 0; (i == 0 && buildFast) || (i < len(potentials) && !buildFast); i++ {
			/* Create new group with this new node */
			newGroup, err := cloneAndAddToGroup(nodeDb, group, potentials[i])
			if err != nil {
				logMessage("Error in cloneAndAddToGroup...")
			} else if groupIsSubsetOfExisting(nodeDb, groupDb, newGroup, false) {
				/*
				 * This group either already exists or a larger group exists
				 * that this is a sub-group of.
				 */
			} else {
				logMessage(fmt.Sprintf("  trying to add node %s", nodeDb.nodes[potentials[i]].externalName))
				if expandGroup(nodeDb, groupDb, newGroup, paramTable, buildFast, minSize) {
					/* found a larger group, making this a subset, so don't add it */
				} else {
					/*
					 * This group does not yet exist and could not be expanded.
					 * So, add it to the GroupDb.  You would think we should only be
					 * adding groups >= minSize.  However, it surprisingly runs
					 * faster with the unnecessary small groups in the GroupDb (which
					 * we just don't print when we're done).
					 */
					addGroup(nodeDb, groupDb, newGroup, true)
					logMessage(fmt.Sprintf("  done building, new size=%d", len(newGroup.ids)))
				}
			}
		}
		/* tell caller we were able to enlarge this group */
		ret = true
	}
	return ret
}

// Sort a list of potential nodes we are considering adding to a group
// by the number of connections each has to the existing group.
func sortPotentials(db *NodeDb, group *Group, potentials []int) []int {
	if len(potentials) == 0 {
		return potentials
	}
	maxCnt := countLinksFromGroupToNode(db, group, potentials[0])

	// nTop will be index into potentials of the node with the most connections
	// to the group.
	nTop := 1
	for i := 1; i < len(potentials); i++ {
		if countLinksFromGroupToNode(db, group, potentials[i]) == maxCnt {
			nTop++
		} else {
			break
		}
	}

	/* If the top contender has more links that any other, no sorting to do */
	if nTop <= 1 {
		return potentials
	}

	/* Create a temporary group of the front of the potentials list of
	 * all the nodes that have the same number of links to our group.
	 * This will help us put the node that not only links the most
	 * to the group in front, but it also links the most to other
	 * nodes we are about to add.
	 */
	var tempGroup Group
	tempGroup.ids = potentials[:nTop]
	nodeCount := make([]NodeCount, nTop)

	for i := 0; i < nTop; i++ {
		nodeCount[i].id = potentials[i]
		nodeCount[i].cnt = countLinksFromGroupToNode(db, &tempGroup, potentials[i])
	}

	/* Sort the node based on how many times that link to our temp group */
	sort.Slice(nodeCount[:], func(i, j int) bool {
		return cmpNodeCountLinkCounts(nodeCount[i], nodeCount[j])
	})

	/* Use the sort results to rearrange our potentials list. */
	for i := 0; i < nTop; i++ {
		potentials[i] = nodeCount[i].id
	}
	return potentials
}
