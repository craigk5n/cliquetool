package main

import (
	"fmt"
)

// Kick out any nodes in a group that don't meet requirements of
// the parameter table.
func pruneGroup(db *NodeDb, group *Group, paramTable *GroupParameters) bool {
	modified := false

	logMessage(fmt.Sprintf("Pruning group %s (%d nodes)", group.groupName, len(group.ids)))

	for done := false; !done; {
		minLinks := minLinksForGroupSize(paramTable, len(group.ids))
		logMessage(fmt.Sprintf("  min links = %d for group size %d", minLinks, len(group.ids)))
		done = true
		for i := 0; i < len(group.ids); i++ {
			nodeId := group.ids[i]
			groupLinks := countLinksFromGroupToNode(db, group, nodeId)
			if groupLinks < minLinks {
				/* Not enough links for this node.  Boot 'em. */
				logMessage(fmt.Sprintf("  removing node %s (%d<%d links)",
					db.nodes[nodeId].externalName, groupLinks, minLinks))
				removeNodeFromGroup(db, group, nodeId)
				/* re-check after removing this node */
				done = false
				modified = true
			}
		}
	}

	logMessage(fmt.Sprintf("  final group size: %d", len(group.ids)))

	updateGroupLinkCounts(db, group)

	return modified
}

func removeBogusGroups(db *NodeDb, groupDb *GroupDb, minSize int, verbose int) {
	toosmall := make([]bool, len(groupDb.groups))
	nTooSmall := 0
	for i := 0; i < len(groupDb.groups); i++ {
		if len(groupDb.groups[i].ids) < minSize {
			toosmall[i] = true
			nTooSmall++
		}
	}

	logMessage(fmt.Sprintf("found %d groups smaller than %d", nTooSmall, minSize))

	newNum := len(groupDb.groups) - nTooSmall
	newGroups := make([]Group, newNum)

	j := 0
	for i := 0; i < len(groupDb.groups); i++ {
		if !toosmall[i] {
			newGroups[j] = groupDb.groups[i]
			j++
		}
	}
	logMessage(fmt.Sprintf("Done removing bogus (<%d) groups: %d -> %d groups",
		minSize, len(groupDb.groups), newNum))

	groupDb.groups = newGroups
}

// Examine all groups and trim them down to meet parameter table
// requirements.
func pruneGroups(db *NodeDb, groups *GroupDb, paramTable *GroupParameters, minSize int, verbose int) {
	nModified := 0
	logMessage("Starting group prune process")

	/*
	 * Prune all groups of minSize or larger.  No point in pruning
	 * groups that are already too small.
	 */
	for i := 0; i < len(groups.groups); i++ {
		if len(groups.groups[i].ids) >= minSize {
			if pruneGroup(db, &groups.groups[i], paramTable) {
				nModified++
			}
		}

		removeBogusGroups(db, groups, minSize, verbose)

		if verbose > 0 {
			fmt.Printf("done pruning groups (%d modified).\n", nModified)
		}
	}
}
