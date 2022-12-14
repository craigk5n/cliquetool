package main

import (
	"fmt"
	"math"
	"os"
	"sort"
)

const (
	ReportFormatMarkdown int = 1
	ReportFormatHtml     int = 2
)

const markdownStart string = "# Cliquetool Group Verification Results\n\n"
const htmlStart string = "<html><body><h1>Cliquetool Group Verification Results</h1>\n"
const DegreeCutoff float64 = 50.0 // 50%
const GroupOverlapPercentError float64 = 80.0

type GroupCount struct {
	group *Group
	cnt   int
}
type GroupReportErrorCount struct {
	reportText string
	errorSize  int
	groupSize  int
}

func linkCountForNode(db *NodeDb, id int) int {
	if id < len(db.nodes) {
		return len(db.nodes[id].links)
	} else {
		return 0
	}
}

// Function callback for sort.Slie on a group error report:
// sort from highest to lowest
func cmpGroupReportErrors(a GroupReportErrorCount, b GroupReportErrorCount) bool {
	if a.errorSize < b.errorSize {
		return true
	} else if a.errorSize > b.errorSize {
		return false
	}
	return true
}

// Function callback for sort.Slize on a group error report:
// sort from highest to lowest
func cmpGroupReportSizes(a GroupReportErrorCount, b GroupReportErrorCount) bool {
	if a.groupSize < b.groupSize {
		return false
	} else if a.groupSize > b.groupSize {
		return true
	}
	return false
}

// function callback for sort.Slize sort from highest to lowest
func cmpGroupCounts(a *GroupCount, b *GroupCount) bool {
	perA := float64(a.cnt) / float64(len(a.group.ids))
	perB := float64(b.cnt) / float64(len(b.group.ids))

	if perA < perB {
		return true
	} else if perA > perB {
		return false
	} else {
		// secondary sort is number of links
		if a.cnt < b.cnt {
			return true
		} else if a.cnt > b.cnt {
			return false
		}
	}
	return false
}

// Add the link into the existing NodeCount array.
func addGroupToGroupCount(groupCount []GroupCount, group *Group, grpCount int) []GroupCount {
	found := false
	cnt := len(groupCount)
	ret := groupCount

	/* search for group in existing array */
	for j := 0; j < cnt && !found; j++ {
		if ret[j].group.groupName == group.groupName {
			ret[j].cnt += grpCount
			found = true
		}
	}
	if !found {
		gc := GroupCount{group: group, cnt: grpCount}
		ret = append(ret, gc)
	}
	return ret
}

// Verify the groups found for a set of nodes and generate a report in either
// markdown or html.
func verifyGroupDb(nodeDb *NodeDb, groupDb *GroupDb, reportOutputPath string, reportFormat int, reportSortBySize bool) error {
	var f *os.File
	if len(reportOutputPath) == 0 {
		f = os.Stdout
	} else {
		var err error
		f, err = os.Create(reportOutputPath)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	if reportFormat == ReportFormatMarkdown {
		f.WriteString(markdownStart)
	} else {
		f.WriteString(htmlStart)
	}

	/* update node group counts */
	updateNodeGroupCounts(nodeDb, groupDb)

	writeGroupSummaryReport(nodeDb, groupDb, f, reportFormat)

	reports := make([]GroupReportErrorCount, 0)

	for i := 0; i < len(groupDb.groups); i++ {
		ret := verifyGroupStruct(nodeDb, groupDb, &groupDb.groups[i], f, reportFormat)
		//if ret.errorSize > 0 {
		/* add to our array */
		reports = append(reports, *ret)
		//}
	}

	if reportSortBySize {
		// sort results, showing largest groups first
		sort.Slice(reports[:], func(i, j int) bool {
			return cmpGroupReportSizes(reports[i], reports[j])
		})
	} else {
		// sort results, showing questionable groups first
		sort.Slice(reports[:], func(i, j int) bool {
			return cmpGroupReportErrors(reports[i], reports[j])
		})
	}

	/* now write to file the sorted group reports */
	for i := 0; i < len(reports); i++ {
		f.WriteString(reports[i].reportText)
	}
	if reportFormat == ReportFormatMarkdown {
		f.WriteString(fmt.Sprintf("\n\nGenerated by cliquetool v%s\n", CliqueToolVersion))
	} else {
		f.WriteString(fmt.Sprintf("\n<hr/><p>Generated by cliquetool v%s\n", CliqueToolVersion))
	}

	return nil

}

// Write a report of group size count
func writeGroupSummaryReport(nodeDb *NodeDb, groupDb *GroupDb, fp *os.File, reportFormat int) {
	/* update group stats (if needed) */
	maxSize := 0
	for i := 0; i < len(groupDb.groups); i++ {
		updateGroupStats(nodeDb, &groupDb.groups[i])
		maxSize = maxOfTwoInts(maxSize, len(groupDb.groups[i].ids))
	}

	cnt := make([]int, maxSize+1)
	maxCnt := 0
	for i := 0; i < len(groupDb.groups); i++ {
		l := len(groupDb.groups[i].ids)
		cnt[l]++
		maxCnt = maxOfTwoInts(maxCnt, l)
	}

	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "## Group Size Summary\n")
		fmt.Fprintf(fp, "| %15s | %15s | %15s |\n", "Group Size", "Count (=)", "Count (>=)")
		fmt.Fprintf(fp, "| %15s | %15s | %15s |\n", "---", "---:", "---:")
	} else {
		fmt.Fprintf(fp, "<table><thead><tr><th colspan\"4\">Group Size Summary</th></tr>\n")
		fmt.Fprintf(fp, "<tr><th>%s</th><th>%s</th><th>%s</th></tr></thead>\n",
			"Group Size", "Count(=)", "Count(>=)")
	}

	sum := 0
	for i := maxSize; i >= 3; i-- {
		if cnt[i] > 0 {
			sum += cnt[i]
			if reportFormat == ReportFormatMarkdown {
				fmt.Fprintf(fp, "| %15d | %15d | %15d |\n", i, cnt[i], sum)
			} else {
				fmt.Fprintf(fp, "<tr><td>%10d</td><td>%10d</td><td>%10d</td></tr>\n", i, cnt[i], sum)
			}
		}
	}

	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "\n")
	} else {
		fmt.Fprintf(fp, "</table>\n<hr>\n")
	}

	// How many valid nodes (some Id numbers may have been skipped)
	totalNumNodes := 0
	for i := 0; i < len(nodeDb.nodes); i++ {
		if len(nodeDb.nodes[i].links) > 0 {
			totalNumNodes++
		}
	}

	// Generate a table that shows a density distribution.  For example,
	// how many groups in the 0 - 10% range and so on.
	//
	densities := make([]int, 11)
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "## Group Density Summary\n")
		fmt.Fprintf(fp, "| %15s | %15s | %15s |\n", "Group Density", "Count (=)", "Count (>=)")
		fmt.Fprintf(fp, "| %15s | %15s | %15s |\n", "---", "---:", "---:")
	} else {
		fmt.Fprintf(fp, "<table><thead><tr><th colspan\"3\">Group Density Summary</th></tr>\n")
		fmt.Fprintf(fp, "<tr><th>%s</th><th>%s</th><th>%s</th></tr></thead>\n",
			"Group Density", "Count(=)", "Count(>=)")
	}
	// count up the number for each bucket.
	// 0 = (0.00 - 0.099), 1 = (0.10-0.199)
	maxCnt = 0
	for i := 0; i < len(groupDb.groups); i++ {
		binNum := int(math.Floor(groupDb.groups[i].density * 10.0))
		if binNum >= 0 && binNum <= 10 {
			densities[binNum]++
			if densities[binNum] > maxCnt {
				maxCnt = densities[binNum]
			}
		}
	}
	sum = 0
	for i := 10; i >= 0; i-- {
		sum += densities[i]
		var temp string
		if i < 10 {
			temp = fmt.Sprintf("%d0.0-%d9.9%%", i, i)
		} else {
			temp = "100%"
		}
		if reportFormat == ReportFormatMarkdown {
			fmt.Fprintf(fp, "| %15s | %15d | %15d |\n", temp, densities[i], sum)
		} else {
			fmt.Fprintf(fp, "<tr><td>%s</td><td>%d</td><td>%d</td></tr>\n",
				temp, densities[i], sum)
		}
	}
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "\n")
	} else {
		fmt.Fprintf(fp, "</table>\n<hr/>\n")
	}

	// Now generate a table that shows how many groups each node is in.
	// (80% are in 1 group, 10% are in 2 groups, etc.)
	maxGroupCount := 0
	for i := 0; i < len(nodeDb.nodes); i++ {
		if nodeDb.nodes[i].numGroups >= maxGroupCount {
			maxGroupCount = nodeDb.nodes[i].numGroups
		}
	}
	groupSizeCount := make([]int, maxGroupCount+1)
	for i := 0; i < len(nodeDb.nodes); i++ {
		if len(nodeDb.nodes[i].links) > 0 {
			numGroupsForNode := nodeDb.nodes[i].numGroups
			groupSizeCount[numGroupsForNode]++
		}
	}
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "## Node Group Membership Counts\n")
		fmt.Fprintf(fp, "| %15s | %15s |\n", "No. Groups", "No. Nodes")
		fmt.Fprintf(fp, "| %15s | %15s |\n", "---:", "---:")
		temp := fmt.Sprintf(" %d (%.2f%%)",
			(totalNumNodes - groupSizeCount[0]),
			100.0*(float64(totalNumNodes)-
				float64(groupSizeCount[0]))/float64(totalNumNodes))
		fmt.Fprintf(fp, "| %15s | %15s |\n", "1+ Groups", temp)
		temp = fmt.Sprintf("%d (%.2f%%)",
			groupSizeCount[0],
			100.0*float64(groupSizeCount[0])/float64(totalNumNodes))
		fmt.Fprintf(fp, "| %15s | %15s |\n", "No Groups", temp)
	} else {
		fmt.Fprintf(fp, "<table><thead><tr><th colspan\"2\">Node Group Membership Counts</th></tr>\n")
		fmt.Fprintf(fp, "<tr><th>%s</th><th>%s</th></tr></thead>\n",
			"No. Groups", "No. Nodes")
		fmt.Fprintf(fp, "<tr class=\"totals\"><td>1 or More Groups</td><td>%d (%.2f%%)</td></tr>\n",
			(totalNumNodes - groupSizeCount[0]),
			100.0*(float64(totalNumNodes)-
				float64(groupSizeCount[0]))/float64(totalNumNodes))
		fmt.Fprintf(fp, "<tr class=\"totals\"><td>No Groups</td><td>%d (%.2f%%)</td></tr>\n",
			groupSizeCount[0],
			100.0*float64(groupSizeCount[0])/float64(totalNumNodes))
	}
	for i := 1; i < maxGroupCount || i < 3; i++ {
		if i > maxGroupCount || i > len(groupSizeCount)-1 {
			if reportFormat == ReportFormatMarkdown {
				fmt.Fprintf(fp, "| %15d | %15s |\n", i, "0 (0.00%)")
			} else {
				fmt.Fprintf(fp, "<tr><td>%d</td><td>%d (%.2f%%)</td></tr>\n",
					i, 0, 0.0)
			}
		} else if groupSizeCount[i] > 0 {
			if reportFormat == ReportFormatMarkdown {
				temp := fmt.Sprintf("%d (%.2f%%)",
					groupSizeCount[i],
					100.0*float64(groupSizeCount[i])/float64(totalNumNodes))
				fmt.Fprintf(fp, "| %15d | %15s |\n", i, temp)
			} else {
				fmt.Fprintf(fp, "<tr><td>%d</td><td>%d (%.2f%%)</td></tr>\n",
					i, groupSizeCount[i], 100.0*float64(groupSizeCount[i])/
						float64(totalNumNodes))
			}
		}
	}
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "\n")
	} else {
		fmt.Fprintf(fp, "</table>\n\n")
	}

	// Print a table of the nodes with the most links that did not
	// end up in any group.
	nodeCount := make([]NodeCount, 0)
	nNodes := 0
	for i := 0; i < len(nodeDb.nodes); i++ {
		if len(nodeDb.nodes[i].links) >= 4 && nodeDb.nodes[i].numGroups == 0 {
			nc := NodeCount{id: nodeDb.nodes[i].id, cnt: len(nodeDb.nodes[i].links)}
			nodeCount = append(nodeCount, nc)
			nNodes++
		}
	}
	sort.Slice(nodeCount[:], func(i, j int) bool {
		return cmpNodeCountLinkCounts(nodeCount[i], nodeCount[j])
	})

	/* show the top 40 or all with a node count of 4 or more... */
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "## Ungrouped Nodes\n")
		if nNodes == 0 {
			fmt.Fprintf(fp, "None\n")
		} else {
			fmt.Fprintf(fp, "| %30s | %15s |\n", "NodeId", "No. Links")
			fmt.Fprintf(fp, "| %30s | %15s |\n", "---:", "---:")
		}
	} else {
		if nNodes == 0 {
			fmt.Fprintf(fp, "None\n")
		} else {
			fmt.Fprintf(fp, "<table class=\"ungroupednodes\"><thead><tr><th colspan\"2\">Ungrouped Nodes</th></tr>\n")
			fmt.Fprintf(fp, "<tr><th>%s</th><th>%s</th></tr></thead>\n",
				"NodeId", "No. Links")
		}
	}
	if nNodes > 0 {
		for i := 0; i < nNodes && i < 40; i++ {
			//don't show nodes with less than 4 links
			if nodeCount[i].cnt < 4 && i > 20 {
				break
			}
			if reportFormat == ReportFormatMarkdown {
				fmt.Fprintf(fp, "| %30s | %15d |\n",
					nodeNameAbbr(nodeDb.nodes[nodeCount[i].id], 30),
					nodeCount[i].cnt)
			} else {
				fmt.Fprintf(fp, "<tr><td>%s</td><td>%d</td></tr>\n",
					nodeDb.nodes[nodeCount[i].id].externalName, nodeCount[i].cnt)
			}
		}
		if reportFormat == ReportFormatMarkdown {
			fmt.Fprintf(fp, "\n")
		} else {
			fmt.Fprintf(fp, "</table>\n\n")
		}
	}

	// Print up a table that shows a one-line summary of each group
	// that includes: name, size, group links, all links for nodes in group,
	// density.
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "## Group Statistics\n")
		fmt.Fprintf(fp, "| %25s | %8s | %12s | %12s | %9s |\n", "Name", "Size", "Group Links",
			"Total Links", "Density")
		fmt.Fprintf(fp, "| %25s | %8s | %12s | %12s | %9s |\n", "---", "---:", "---:", "---:", "---:")
	} else {
		fmt.Fprintf(fp, "<table><thead><tr><th colspan\"5\">Group Statistics</th></tr>\n")
		fmt.Fprintf(fp, "<tr><th>%s</th><th>%s</th><th>%s</th></tr></thead>",
			"Name", "Size", "Group Links")
	}
	for i := 0; i < len(groupDb.groups); i++ {
		group := &groupDb.groups[i]
		if reportFormat == ReportFormatMarkdown {
			fmt.Fprintf(fp, "| %25s | %8d | %12d | %12d | %9.2f |\n", group.groupName, len(group.ids),
				group.groupLinks,
				group.totalNodeLinks, 100.0*group.density)
		} else {
			fmt.Fprintf(fp, "<tr><td>%s</td><td>%d</td><td>%d</td><td>%d</td><td>%0.2f%%</td></tr>\n",
				group.groupName, len(group.ids),
				group.groupLinks, group.totalNodeLinks, 100.0*group.density)
		}
	}
	if reportFormat == ReportFormatMarkdown {
		fmt.Fprintf(fp, "\n")
	} else {
		fmt.Fprintf(fp, "</table>\n\n")
	}
}

// Run some verification tools and gather some statistics about a group.
// If the groupDb parameter is non-nil, then we will also check
// to see if the group is related in any way to other groups
// (subset, superset, overlap, etc.)
func verifyGroupStruct(db *NodeDb, groupDb *GroupDb, group *Group,
	fp *os.File, reportFormat int) *GroupReportErrorCount {
	totalLinkCnt := 0
	minLinksInGroup := -1 // min inner-group links for any node
	maxMissing := 10
	externalAve := 0.0
	var ret GroupReportErrorCount

	ret.groupSize = len(group.ids)
	if reportFormat == ReportFormatMarkdown {
		ret.reportText += fmt.Sprintf("## Summary For Group: %s\n", group.groupName)
		ret.reportText += fmt.Sprintf("| %25s | %14s | %16s | %10s |\n", "NodeID", "Group Links",
			"Grp/Total Lnks", "No. Groups")
		ret.reportText += fmt.Sprintf("| %25s | %14s | %16s | %10s |\n", "---", "---:", "---:", "---:")
	} else {
		ret.reportText += fmt.Sprintf("<table><thead><tr><th colspan\"6\">Group %s</th></tr>\n", group.groupName)
		ret.reportText += fmt.Sprintf("<tr><th>%s</th><th colspan=\"2\">%s</th><th colspan=\"2\">%s</th><th>%s</th></tr></thead>",
			"NodeID", "Group Links", "Grp/Total Lnks", "No. Groups")
	}

	for i := 0; i < len(group.ids); i++ {
		// count how many in this group this node links to
		linkCnt := 0
		allNodeLinkCnt := linkCountForNode(db, group.ids[i])
		for j := 0; j < len(group.ids); j++ {
			if i != j {
				if nodeLinksTo(*db, group.ids[i], group.ids[j]) {
					linkCnt++
				}
			}
		}
		if linkCnt < minLinksInGroup || minLinksInGroup < 0 {
			minLinksInGroup = linkCnt
		}
		totalLinkCnt += linkCnt
		temp := fmt.Sprintf("%d/%d %6.2f%%", linkCnt, len(group.ids)-1,
			100.0*float64(linkCnt)/float64(len(group.ids)-1))
		temp2 := fmt.Sprintf("%d/%d %6.2f%%", linkCnt, allNodeLinkCnt,
			100.0*float64(linkCnt)/float64(allNodeLinkCnt))
		if reportFormat == ReportFormatMarkdown {
			line := fmt.Sprintf("| %25s | %14s | %16s | %10d\n",
				nodeNameAbbr(db.nodes[group.ids[i]], 18), temp, temp2,
				db.nodes[group.ids[i]].numGroups)
			ret.reportText += line
		} else {
			line := fmt.Sprintf(
				"<tr><td>%s</td>%s%s<td>%d</td></tr>\n",
				db.nodes[group.ids[i]].externalName,
				temp, temp2, db.nodes[group.ids[i]].numGroups)
			ret.reportText += line
		}
		if linkCnt == 0 {
			ret.errorSize = 100
		}
		externalAve += 100.0 *
			float64(linkCnt) / float64(allNodeLinkCnt)
	}

	groupAve := 100.0 * float64(totalLinkCnt) /
		float64((len(group.ids)-1)*len(group.ids))
	var line string
	if reportFormat == ReportFormatMarkdown {
		temp := fmt.Sprintf("%d/%d %6.2f%%", totalLinkCnt,
			(len(group.ids)-1)*len(group.ids), groupAve)
		temp2 := fmt.Sprintf("AVE %6.2f%%", externalAve/float64(len(group.ids)))
		line = fmt.Sprintf("| %25s | %14s | %16s | %10s | \n", "TOTALS", temp, temp2, "-")
	} else {
		temp := fmt.Sprintf("<td>%d/%d</td><td>%.2f%%</td>", totalLinkCnt,
			(len(group.ids)-1)*len(group.ids), groupAve)
		temp2 := fmt.Sprintf("<td>AVE</td><td>%.2f%%</td>",
			externalAve/float64(len(group.ids)))
		line = fmt.Sprintf("<tr class=\"totals\"><td>%s</td>%s%s<td>-</td></tr>\n",
			"TOTALS", temp, temp2)
	}
	ret.reportText += line

	/*
	 * Now, we want to find if there are any possible group members that
	 * might have been missed.  We do this by counting up all the links
	 * that the group members have.  Then, we sort the list to find
	 * out which non-group member is linked to the most.
	 * If any of these have more than the least-linked within the
	 * group, we mark it with either a '!' or '+'.
	 */
	nodeCount := make([]NodeCount, 0)
	for i := 0; i < len(group.ids); i++ {
		/* get all links for node id group.ids[i] */
		nodeId := group.ids[i]
		for j := 0; j < len(db.nodes[nodeId].links); j++ {
			/* Make sure this link does not reference a current group node */
			if !isInArray(group.ids, db.nodes[nodeId].links[j]) {
				nodeCount = addLinkToNodeCount(db, nodeCount, db.nodes[nodeId].links[j])
			}
		}
	}
	if len(nodeCount) == 0 {
		if reportFormat == ReportFormatMarkdown {
			ret.reportText += "### No other possible group nodes found\n"
		} else {
			ret.reportText += "<tr><td colspan=\"6\">No other possible group nodes found</td></tr>\n"
		}
	} else {
		// Sort the nodes by number of links to it in the group.
		// Secondary sort is the number of total links (less is better).
		sort.Slice(nodeCount[:], func(i, j int) bool {
			return cmpNodeCountLinkCounts(nodeCount[i], nodeCount[j])
		})
		// print either top 5 or all that have more than min links
		if reportFormat == ReportFormatMarkdown {
			ret.reportText += "### Potential group nodes not included\n" +
				fmt.Sprintf("| %25s | %14s | %16s | %10s |\n", "NodeID", "Group Links",
					"Grp/Total Lnks", "No. Groups") +
				fmt.Sprintf("| %25s | %14s | %16s | %10s |\n", "---", "---:", "---:", "---:")
		} else {
			ret.reportText += "<tr><td colspan=\"6\">Potential group nodes not included</td></tr>\n"
		}
		for i := 0; i < len(nodeCount); i++ {
			ch := ' '
			// count how many in this group this node links to
			linkCnt := 0
			allNodeLinkCnt := linkCountForNode(db, nodeCount[i].id)
			for j := 0; j < len(group.ids); j++ {
				if nodeLinksTo(*db, nodeCount[i].id, group.ids[j]) {
					linkCnt++
				}
			}
			degree := 100.0 * float64(linkCnt) / float64(len(group.ids))
			//ieRatio := 100.0 * float64(linkCnt) / float64(allNodeLinkCnt)
			// Show node and continue if any of the following:
			// - degree is >= DEGREE_CUTOFF
			// - this is one of the first 5 nodes in the list
			if i >= 5 && degree < DegreeCutoff {
				continue
			}
			// Only show 10 candidates if we are still showing nodes that SHOULD
			// be considered for inclusion in the group */
			if linkCnt >= minLinksInGroup && i >= maxMissing {
				if reportFormat == ReportFormatMarkdown {
					ret.reportText += "| more... |\n"
				} else {
					ret.reportText += "<tr><td colspan=\"6\" class=\"str\">[more...]</td></tr>\n"
				}
				break
			}
			// Use some special characters to make searching through the output
			// for important stuff a little easier.
			// Use '+' to denote that this entry is linked to more group nodes
			// than the least-linked group member is.
			// Use '!' to denote the above AND the ratio of links to the group
			// to total group nodes is higher than the average for the whole
			// group.  (Seems like this shouldn't happen to me...)
			if linkCnt >= minLinksInGroup && degree >= groupAve {
				ret.errorSize = linkCnt - minLinksInGroup
				ch = '!'
			} else if linkCnt >= minLinksInGroup {
				ch = '+'
			} else if degree >= DegreeCutoff {
				// TODO: replace DegreeCutoff with value from ParamTable
				ch = '+'
			}
			if reportFormat == ReportFormatMarkdown {
				temp := fmt.Sprintf("%d/%d %6.2f%%", linkCnt, len(group.ids), degree)
				temp2 := fmt.Sprintf("%d/%d %6.2f%%", linkCnt, allNodeLinkCnt,
					100.0*float64(linkCnt)/float64(allNodeLinkCnt))
				ret.reportText +=
					fmt.Sprintf("| %c%24s | %14s | %16s | %10d |\n", ch,
						nodeNameAbbr(db.nodes[nodeCount[i].id], 17), temp, temp2,
						db.nodes[nodeCount[i].id].numGroups)
			} else {
				temp := fmt.Sprintf("<td>%d/%d</td><td>%.2f%%</td>", linkCnt,
					len(group.ids), degree)
				temp2 := fmt.Sprintf("<td>%d/%d</td><td>%.2f%%</td>", linkCnt,
					allNodeLinkCnt,
					100.0*float64(linkCnt)/float64(allNodeLinkCnt))
				ret.reportText +=
					fmt.Sprintf("<tr><td>%s</td>%s%s<td>%d</td></tr>\n",
						db.nodes[nodeCount[i].id].externalName, temp, temp2,
						db.nodes[nodeCount[i].id].numGroups)
			}
		}
	}
	// See if this group is connected (subset, superset, overlap) to
	// any groups specified in the GroupDb passed in.
	// The GroupDb passed may or may not include this group, so we
	// may find this exact group in the GroupDb.
	if len(nodeCount) > 0 {
		groupCount := make([]GroupCount, 0)
		for i := 0; i < len(groupDb.groups) && groupDb.groups[i].groupName != group.groupName; i++ {
			group2 := &groupDb.groups[i]
			j := countSharedNodes(group, group2)
			// Technically, j=1 would be one common member in two groups, but
			// this generates too much info to look at.
			if j >= 3 {
				// If we are comparing groups within the same group db, don't
				// compare a group against itself.
				if group.groupName == group2.groupName {
					groupCount =
						addGroupToGroupCount(groupCount, group2, j)
				}
			}
		}
		if len(groupCount) > 0 {
			numSupergroups := 0
			numRelated := 0
			// now sort groupCount to put groups with most overlap first
			sort.Slice(groupCount[:], func(i, j int) bool {
				return cmpGroupCounts(&groupCount[i], &groupCount[j])
			})
			for i := 0; i < len(groupCount); i++ {
				foundSuspectSubset := false
				if groupCount[i].cnt == len(group.ids) {
					if numSupergroups == 0 {
						if reportFormat == ReportFormatMarkdown {
							ret.reportText += "### This group is a subset of:\n" +
								fmt.Sprintf("| %25s | %14s |\n", "---:", "---") +
								fmt.Sprintf("| %25s | %14s |\n", "Shared/Total", "Group")
						} else {
							ret.reportText += "<tr><th colspan=\"6\">This group is a subset of</th></tr>\n" +
								"<tr><th>Shared/Total</th><th colspan=\"5\">Group</th></tr>\n" +
								fmt.Sprintf("<tr><td>%18s</td><td>%s</td></tr>\n", "Shared/Total", "Group")
						}
					}
					numSupergroups++
					if !foundSuspectSubset {
						ret.errorSize += 100
						foundSuspectSubset = true
					}
				} else {
					if !foundSuspectSubset &&
						groupCount[i].cnt == len(groupCount[i].group.ids) {
						ret.errorSize += 100
						foundSuspectSubset = true
					}
					if numRelated == 0 {
						if reportFormat == ReportFormatMarkdown {
							ret.reportText += "### Related groups (3+ common nodes):\n" +
								fmt.Sprintf("| %25s | %14s |\n", "---:", "---") +
								fmt.Sprintf("| %25s | %14s |\n", "Shared/Total", "Group")
						} else {
							ret.reportText += "<tr><th colspan=\"6\">Related groups (3+ common nodes)</th></tr>\n" +
								fmt.Sprintf("<tr><th>%s</th><th colspan=\"5\">%s</th></tr>\n",
									"Shared/Total", "Group")
						}
					}
					numRelated++
				}
				temp := fmt.Sprintf("%d/%d (%5.2f%%)", groupCount[i].cnt,
					len(groupCount[i].group.ids),
					100.0*float64(groupCount[i].cnt)/
						float64(len(groupCount[i].group.ids)))
				ch := ' '
				if len(groupCount[i].group.ids) == groupCount[i].cnt {
					ch = '='
				}
				if reportFormat == ReportFormatMarkdown {
					ret.reportText += fmt.Sprintf("| %c%24s | %14s |\n", ch,
						temp, groupCount[i].group.groupName)
				} else {
					ret.reportText += fmt.Sprintf("<tr>><td>%c%s</td><td colspan=\"5\" class=\"str\">%s</td></tr>\n",
						ch, temp, groupCount[i].group.groupName)
				}
				if 100.0*float64(groupCount[i].cnt)/
					float64(len(groupCount[i].group.ids)) >=
					GroupOverlapPercentError {
					/* found another group that is a subset of this one */
					ret.errorSize++
				}
				if i == 9 && len(groupCount) >= 10 {
					/* only display 10... */
					if reportFormat == ReportFormatMarkdown {
						ret.reportText += "[more...]\n"
					} else {
						ret.reportText += "<tr><td colspan=\"6\" class=\"str\">[more...]</td></tr>\n"
					}
					break
				}
			}
		} else {
			if reportFormat == ReportFormatMarkdown {
				ret.reportText += "No related groups\n"
			} else {
				ret.reportText += "<tr><td colspan=\"6\" class=\"str\">No related groups</td></tr>\n"
			}
		}
	}

	if reportFormat == ReportFormatMarkdown {
		ret.reportText += "\n"
	} else {
		ret.reportText += "</table><br/><br/>\n\n"
	}
	return &ret
}
