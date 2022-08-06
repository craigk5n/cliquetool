package main

import (
	"fmt"
	"os"
)

func dumpGroups(nodeDb *NodeDb, groupDb *GroupDb) {
	fmt.Println("** GroupDb **")
	for i := 0; i < len(groupDb.groups); i++ {
		dumpGroup(nodeDb, &groupDb.groups[i], i)
	}
	fmt.Println("**")
}
func dumpGroup(nodeDb *NodeDb, group *Group, num int) {
	fmt.Printf("Group #%d: name='%s', size=%d links=", num, group.groupName, len(group.ids))
	for j := 0; j < len(group.ids); j++ {
		if j > 0 {
			fmt.Print(",")
		}
		fmt.Print(group.ids[j])
	}
	fmt.Println("")
}

func writeGroupFile(nodeDb *NodeDb, groupDb *GroupDb, path string, minGroupSize int, verbose int) error {
	// TODO: forbid overwrite?
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("group,member\n\n")
	cnt := 0
	largest := 0
	for _, group := range groupDb.groups {
		if len(group.ids) >= minGroupSize {
			if verbose > 1 {
				fmt.Printf("Processing group: %s with %d links\n", group.groupName, len(group.ids))
			}
			for _, nodeID := range group.ids {
				if verbose > 1 {
					fmt.Println("My node:", group.groupName, nodeID)
				}
				f.WriteString(fmt.Sprintf("%s,%s\n", group.groupName, nodeDb.nodes[nodeID].externalName))
			}
			cnt++
		}
		if len(group.ids) > largest {
			largest = len(group.ids)
		}
	}
	fmt.Printf("Wrote %d groups (largest is %d) to %s\n", cnt, largest, path)

	return nil
}
