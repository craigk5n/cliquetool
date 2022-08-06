package main

import (
	"fmt"
	"sort"

	"github.com/schollz/progressbar/v3"
)

const StatusNotMerged int = 0
const StatusMerged int = 1

type GroupOverlapCount struct {
	group *Group
	cnt   int
}

func maxOfTwoInts(int1 int, int2 int) int {
	if int1 > int2 {
		return int1
	} else {
		return int2
	}
}

// Is the specified integer in the array of integers.
func isInArray(ids []int, idToFind int) bool {
	for i := 0; i < len(ids); i++ {
		if ids[i] == idToFind {
			return true
		}
	}
	return false
}

func isInStringArray(arr []string, name string) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == name {
			return true
		}
	}
	return false
}

// Function callback for sort: sort from highest to lowest that will
// return true if the first parameter has a higher overlap percentage
// than the second parameter.
func cmpGroupOverlap(a GroupOverlapCount, b GroupOverlapCount) bool {
	perA := float64(a.cnt) / float64(len(a.group.ids))
	perB := float64(b.cnt) / float64(len(b.group.ids))

	if perA < perB {
		return false
	} else if perA > perB {
		return true
	} else {
		/* secondary sort is number of links */
		if a.cnt < b.cnt {
			return false
		} else if a.cnt > b.cnt {
			return true
		}
		return true
	}
}

func isIdenticalGroup(group1 *Group, group2 *Group) bool {
	if group1.groupName != group2.groupName || len(group1.ids) != len(group2.ids) {
		return false
	}
	if len(group1.linkCounts) != len(group2.linkCounts) || group1.density != group2.density || group1.statsAreCurrent != group2.statsAreCurrent {
		return false
	}
	if group1.status != group2.status || group1.totalNodeLinks != group2.totalNodeLinks {
		return false
	}
	for i := 0; i < len(group1.ids); i++ {
		if group1.ids[i] != group2.ids[i] {
			return false
		}
	}
	for i := 0; i < len(group1.linkCounts); i++ {
		if group1.linkCounts[i] != group2.linkCounts[i] {
			return false
		}
	}
	return true
}

// Add the link into the existing NodeCount array.
func addGroupToGroupOverlapCount(groupCount []GroupOverlapCount, group *Group, grpCount int) []GroupOverlapCount {
	{
		ret := groupCount
		found := false

		// search for group in existing array
		for j := 0; j < len(groupCount) && !found && ret != nil; j++ {
			if isIdenticalGroup(group, groupCount[j].group) {
				/* found it */
				ret[j].cnt += grpCount
				found = true
			}
		}

		if !found {
			/* append to end of list */
			newCount := GroupOverlapCount{group, grpCount}
			ret = append(ret, newCount)
		}
		return ret
	}
}

// Take a set of groups and merge overlapping groups using
// the rules specified in the paramTable parameter.
// Normally, the input groups would be unmerged dense groups, but
// you could also pass in partially merged groups created from another
// link analysis tool.
func mergeGroups(db *NodeDb, origGroups *GroupDb,
	mergeOverlapRatio float64, maxMissingGroupLinks int,
	paramTable *GroupParameters, minSize int, verbose int) *GroupDb {
	logMessage(fmt.Sprintf("Merging groups with %.2f%% overlap...",
		100.0*mergeOverlapRatio))
	if maxMissingGroupLinks > 0 && verboseLevel > 0 {
		logMessage(fmt.Sprintf("  Groups can only have %d missing links",
			maxMissingGroupLinks))
	}

	var ret GroupDb
	/* Make a copy of the original groups so that we do not modify them */
	clonedGroups := cloneGroupDb(origGroups)

	/* Set status of all groups to StatusNotMerged */
	for i := 0; i < len(clonedGroups.groups); i++ {
		clonedGroups.groups[i].status = StatusNotMerged
	}

	/* Sort the cloned groups. */
	sortGroupDb(clonedGroups)

	fmt.Printf("Merging groups.\n")
	bar := progressbar.Default(int64(len(clonedGroups.groups)))
	for i := 0; i < len(clonedGroups.groups); i++ {
		if clonedGroups.groups[i].status == StatusNotMerged {
			attemptToMergeGroup(db, &ret, clonedGroups, &clonedGroups.groups[i],
				mergeOverlapRatio, maxMissingGroupLinks, paramTable)
		}
		updateGroupLinkCounts(db, &clonedGroups.groups[i])
		bar.Add(1)
	}
	bar.Finish()
	bar.Close()

	fmt.Printf("\nDone merging groups: %d -> %d groups\n", len(origGroups.groups),
		len(ret.groups))

	// Remove subsets.  Do this before we prune since we may prune off one
	// node that prevents a group from being a subset.
	removeGroupSubsets(&ret, verbose)

	// sort groups to return
	sortGroupDb(&ret)

	return &ret
}

// Take the specified group and see if it can be merged
// with any other groups.  If so, we will mark it as having
// been merged. Either way we put the results into the newGroups db.
// So, either we put a merged group in newGroups or the unmodified
// group in it.
func attemptToMergeGroup(db *NodeDb,
	newGroups *GroupDb, oldGroups *GroupDb,
	group *Group, mergeOverlapRatio float64,
	maxMissingGroupLinks int, paramTable *GroupParameters) {

	logMessage(fmt.Sprintf("Attempting to merge group: %s", group.groupName))

	var mergedGroup *Group = nil
	mergedGroups := make([]string, 0)
	done := false
	nGroups := 0
	pass := 0
	var overlapCount []GroupOverlapCount = make([]GroupOverlapCount, 0)
	for !done {
		pass++
		logMessage(fmt.Sprintf("  merge pass %d", pass))
		// Find the group that has the most overlap with this one.
		// We will look at groups that have already been merged, too.
		// So, a group can possibly be merged into multiple other groups.
		nGroups = 0
		for i := 0; i < len(oldGroups.groups); i++ {
			group2 := &oldGroups.groups[i]
			var group1 *Group
			if mergedGroup == nil {
				group1 = group
			} else {
				group1 = mergedGroup
			}
			// don't compare a group against itself
			if group1.groupName == group2.groupName {
				continue
			}
			overlap := countSharedNodes(group1, group2)
			// don't merge the same group twice
			if len(mergedGroups) > 0 &&
				isInStringArray(mergedGroups, group2.groupName) {
				continue
			}
			// calculate overlap
			percentOverlap := float64(overlap) / float64(maxOfTwoInts(len(group1.ids), len(group2.ids)))
			if percentOverlap >= mergeOverlapRatio {
				doMerge := true
				// This group meets overlap threshold.  Now check for number of
				// missing links if two groups are merged.  Would be nice not to
				// have to do the merge, but these calculations are actually
				// quicker after we do a merge since all the duplicate ids get
				// removed.
				if maxMissingGroupLinks > 0 {
					newGroup := mergeTwoGroups(db, group1, group2)
					updateGroupStats(db, newGroup)
					maxLink := len(newGroup.ids) * (len(newGroup.ids) - 1)
					missing := maxLink - newGroup.groupLinks
					// keep in mind we count the link in both directions, so there
					// are two links between each linked node
					if missing > 2*maxMissingGroupLinks {
						doMerge = false
					}
				}
				if doMerge {
					overlapCount =
						addGroupToGroupOverlapCount(overlapCount, group2,
							overlap)
				}
			}
		}
		if len(overlapCount) == 0 || nGroups == 0 {
			/* No more candidate groups for merging.  We are done. */
			done = true
			logMessage("  No overlapping groups found")
		} else {
			logMessage(fmt.Sprintf("Found %d overlapping groups", nGroups))
			// now sort overlapCount to put groups with most overlap first
			sort.Slice(overlapCount[:], func(i, j int) bool {
				return cmpGroupOverlap(overlapCount[i], overlapCount[j])
			})
			// Only the first group.
			group2 := overlapCount[0].group
			logMessage(fmt.Sprintf("  Top overlap (%d nodes): %s", overlapCount[0].cnt, group2.groupName))
			var newGroup *Group
			if len(mergedGroups) == 0 {
				newGroup = mergeTwoGroups(db, group, group2)
			} else {
				newGroup = mergeTwoGroups(db, mergedGroup, group2)
			}
			mergedGroup = newGroup
			mergedGroups = append(mergedGroups, group2.groupName)
			// Add merged group to GroupDb
			group.status = StatusMerged
			group2.status = StatusMerged
		}
	}

	if mergedGroup == nil {
		/* no merging done.  just add original */
		if groupIsSubsetOfExisting(db, newGroups, group, false) {
			logMessage("  group is a duplicate... not adding")
		} else {
			logMessage("  adding original group")
			addGroup(db, newGroups, group, false)
		}
	} else {
		/* merging done.  now add */
		if groupIsSubsetOfExisting(db, newGroups, mergedGroup, false) {
			logMessage("  merged group is a duplicate... not adding")
		} else {
			logMessage("  adding merged group")
			addGroup(db, newGroups, mergedGroup, false)
		}
	}

	group.status = StatusMerged
}

// Merge two groups into a new group.
// We don't check whether this group is valid.  We just merge them
// and return the results.
// We assume group1 is the larger group and will give the new
// group the name of group1.
func mergeTwoGroups(db *NodeDb, group1 *Group, group2 *Group) *Group {
	//var ret *Group = nil

	logMessage(fmt.Sprintf("  merging %s (%d) and %s (%d)",
		group1.groupName, len(group1.ids), group2.groupName, len(group2.ids)))

	/* Start by cloning group1 */
	var clone Group = *group1

	copy(clone.ids, group1.ids)
	/* use name of larger group */
	if len(group1.ids) > len(group2.ids) {
		clone.groupName = group1.groupName
	} else {
		clone.groupName = group2.groupName
	}
	/* Now, add group2 to the cloned new group */
	for _, id := range group2.ids {
		// add the node (if it is not already in the group)
		addNodeToGroup(db, &clone, id, false)
	}

	/* now sort nodes in group */
	sort.Ints(clone.ids[:])

	logMessage(fmt.Sprintf("  new group %s had %d nodes",
		clone.groupName, len(clone.ids)))
	return &clone
}
