package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

// DefaultMergeRatioThreshold sets theshold of percent overlapping nodes that force
// two groups to be merged
const DefaultMergeRatioThreshold = 0.6 // 60% shared nodes

// Group defines a group of connected nodes
type Group struct {
	// Unique group name
	groupName string
	// slice of node IDs
	ids []int
	// Link count array, one for each node.  This is the count of links within this group
	linkCounts []int
	// total number of links within the group
	groupLinks int
	// total link for all nodes (non-group links,too)
	totalNodeLinks int
	// density of group (0.0-1.0)
	density float64
	// Are the above stats current? or do they need updating?
	statsAreCurrent bool
	// status (used in merging process)
	status int
	// Mutex to protect access during concurrency
	mutex sync.Mutex
}

// GroupDb defines the internal database of Group objects
type GroupDb struct {
	// Slice of all groups
	groups []Group
	// Mutex to protect access during concurrency
	mutex sync.Mutex
}

// Find MOST groups of minSize or larger.
// This will find only "dense clusters" (aka complete graphs) where
// each node has a link to every other node in the group.
//
// This will find all isolated groups.  Because it does not do an
// exhaustive search (which could take days or longer to run), there
// are some cases where a dense group will not be found.  This can only
// happen when every group member is already a member of at least one
// other group.  In these cases our subgroup detection may prevent
// us from finding such a group.
//
// NOTE: This app was intended to find all the cliques of a certain size
// or larger.  It may be possible to optimize or improve the odds of
// finding the max clique if we toss aside every clique smaller than
// what we know to be the largest clique.  For example, some of the
// DIMACS data tells us the max clique size.  If we toss aside all
// smaller groups, it may speed up the process or avoid missing the
// correct max clique.
func findGroups(db *NodeDb, minSize int, searchPasses int, verbose int, numThreads int) (*GroupDb, error) {
	var groupDb GroupDb
	var validNodes int

	fmt.Printf("Finding cliques...\n")
	startTime := time.Now()

	groupDb.groups = []Group{}

	validNodes = 0
	for i := 0; i < len(db.nodes); i++ {
		if len(db.nodes[i].links) > 0 {
			validNodes++
		}
	}
	if validNodes == 0 {
		return &groupDb, fmt.Errorf("no valid nodes")
	}

	// start group name at 0
	resetGroupNameGenerator()

	bar := progressbar.Default(int64(len(db.nodes)))
	for i := 0; i < len(db.nodes); i++ {
		// If this node has at least the minSize number of links, it is possible */
		// that they are part of a complete/dense group */
		if len(db.nodes[i].links) >= minSize {
			// Create a group with just two nodes for each of the links that
			// this node has.  Then, try to expand each of these groups
			// with our recursive expandGroup function.
			for j := 0; j < len(db.nodes[i].links); j += numThreads {
				var wg sync.WaitGroup
				for k := 0; k < numThreads && j+k < len(db.nodes[i].links); k++ {
					var group Group
					group.groupName = generateGroupName()
					group.ids = []int{db.nodes[i].id, db.nodes[i].links[j+k]}
					wg.Add(1)
					go expandGroup(&wg, db, &groupDb, &group, nil, true, minSize)
				}
				wg.Wait()
			}
		}
		bar.Add(1)
	}
	bar.Finish()
	bar.Close()

	now := time.Now()
	if now != startTime {
		timeToComplete := (now.UnixMilli() - startTime.UnixMilli()) / 1000
		durH := timeToComplete / 3600
		durM := (timeToComplete / 60) % 60
		durS := timeToComplete % 60
		fmt.Printf("\nProcessing time: %02d:%02d:%02d\n", durH, durM, durS)
	}

	if searchPasses != 1 {
		findMissingGroups(db, &groupDb, minSize, searchPasses, verbose)
	}

	// update node group counts
	// TODO: updateNodeGroupCounts(db, groups)

	// Now sort them; most nodes first
	sortGroupDb(&groupDb)

	// now, rename them
	renameGroups(&groupDb, "clique")

	now = time.Now()
	timeToComplete := int32((now.UnixMilli() - startTime.UnixMilli()) / 1000.0)
	durH := timeToComplete / 3600
	durM := (timeToComplete / 60) % 60
	durS := timeToComplete % 60
	fmt.Printf("Total Processing time: %02d:%02d:%02d\n", durH, durM, durS)
	fmt.Printf("Found %d cliques of size %d or larger\n", len(groupDb.groups), minSize)
	if len(groupDb.groups) > 0 {
		fmt.Printf("Max clique size is %d\n", len(groupDb.groups[0].ids))
	}

	return &groupDb, nil
}

// After we have made the first pass at finding cliques
// with findGroups, if we are only interested in cliques,
// we can do some further searching.  If we are going to build
// on these cliques, this step is probably not needed since
// you will end up with the same groups without this step.
//
// Sometimes the findGroups algorithm will miss some cliques.
// This should only happen when every group member is already a member
// of at least one other group.  In these cases our subgroup detection
// may prevent us from finding such a group.
func findMissingGroups(db *NodeDb, groups *GroupDb,
	minSize int, numPasses int, verbose int) {
	oldNumGroups := 0
	origNumGroups := len(groups.groups)
	i := 0
	for i = 2; i <= numPasses || numPasses <= 0; i++ {
		startPos := oldNumGroups
		oldNumGroups = len(groups.groups)
		searchForMissedGroups(db, groups, minSize, startPos, i, verbose)
		if oldNumGroups == len(groups.groups) {
			if verbose > 0 {
				fmt.Printf("No new groups added in pass %d\n", i)
			}
			break
		}
	}

	// If numPasses is 0 or -1, then do one more big recheck since the
	// addition of more groups may have affected groups we didn't recheck
	// originally.
	if numPasses <= 0 && origNumGroups != len(groups.groups) {
		searchForMissedGroups(db, groups, minSize, 0, i+1, verbose)
	}
}

// Search likely places for groups that may have been missed.
// We can skip any group that has a member that only belongs to
// that group.  So, we first count up group memberships for
// all nodes.  Then, we tag any group with all members that belong
// to two or more groups.  We re-examine these groups and often
// find new groups.  In the process of this, more groups become
// eligible for another pass.
func searchForMissedGroups(db *NodeDb, groups *GroupDb,
	minSize int, startPos int, pass int, verbose int) int {
	fmt.Printf("Searching for missed cliques (pass %d)...\n", pass)
	startTime := time.Now()

	startGroupCount := len(groups.groups)
	nodeGroupCount := make([]int, len(db.nodes))

	if verbose > 0 {
		fmt.Printf("Counting node group memberships for groups %d-%d (pass %d)\n",
			startPos, startGroupCount-1, pass)
	}

	for i := 0; i < len(groups.groups); i++ {
		group := &groups.groups[i]
		for j := 0; j < len(group.ids); j++ {
			nodeGroupCount[group.ids[j]]++
		}
	}

	// Loop through groups looking for groups that have members that
	// each belong to one or more others groups.  So, each member will
	// have a group count of 2 or more.
	if verbose > 0 {
		fmt.Printf("Looking for group candidates (pass %d)\n", pass)
	}

	// First, count how many we will look at so we can show progress */
	groupsToCheck := 0
	groupsArray := make([]int, len(groups.groups))
	for i := startPos; i < startGroupCount; i++ {
		group := &groups.groups[i]
		for j := 0; j < len(group.ids); j++ {
			if nodeGroupCount[group.ids[j]] < 2 {
				// This group member is only a member of this group and no others
				groupsArray[i] = 1
				groupsToCheck++
				break
			}
		}
	}
	// Now, do the actually rechecking of the groups
	if groupsToCheck > 0 {
		fmt.Printf("Rechecking %d groups on pass %d...\n", groupsToCheck, pass)
		bar := progressbar.Default(int64(startGroupCount))
		for i := startPos; i < startGroupCount && groupsToCheck > 0; i++ {
			group := &groups.groups[i]
			if groupsArray[i] > 0 {
				// this group's memeber all belong to 2 or more groups
				recheckGroup(db, groups, group, minSize, verbose)
			}
			bar.Add(1)
		}
		bar.Finish()
		now := time.Now()
		if now != startTime {
			timeToComplete := int32((now.UnixMilli() - startTime.UnixMilli()) / 1000.0)
			durH := timeToComplete / 3600
			durM := (timeToComplete / 60) % 60
			durS := timeToComplete % 60
			fmt.Printf("Processing time: %02d:%02d:%02d\n", durH, durM, durS)
		}
	} else {
		fmt.Printf("No groups needed recheck on pass %d.\n", pass)
	}

	fmt.Printf("Found %d missing groups on pass %d.\n",
		len(groups.groups)-startGroupCount, pass)

	// return the number of new groups found
	return (len(groups.groups) - startGroupCount)
}

// Re-examine a single group for possible dense groups that may
// have been missed initially.  This can happen when all group
// members are connected to two or more groups.
func recheckGroup(db *NodeDb, groups *GroupDb, group *Group, minSize int, verbose int) {
	// Try removing each node (one at a time) from the group.
	// Then, see if the resulting smaller group can be expanded.  We
	// cannot use the subgroup check here because that would be why this
	// group was missed in the first place.
	for i := 0; i < len(group.ids); i++ {
		idToRemove := group.ids[i]
		smallerGroup := cloneAndSubtractFromGroup(db, group, idToRemove)
		//n := 0
		potentials := nodesThatLinkToAllOfGroup(db, smallerGroup)
		for j := 0; len(potentials) > 0 && j < len(potentials); j++ {
			idToAdd := potentials[j]
			// must skip Id we just removed (or cloneAndAddToGroup will barf)
			if idToAdd == idToRemove {
				continue
			}
			// Create new group with this new node
			group2, _ := cloneAndAddToGroup(db, smallerGroup, idToAdd)
			//  Make sure we cannot expand further...
			/* group already found */
			if !expandGroup(nil, db, groups, group2, nil, true, minSize) {
				/* addGroup will make sure not to add a subset */
				addGroup(db, groups, group2, true)
			}
		}
	}
}
