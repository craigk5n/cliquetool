package main

import (
	"testing"
)

func TestMaxOfTwoInts(t *testing.T) {
	if maxOfTwoInts(1, 2) != 2 {
		t.Errorf("Max should be 2")
	}
	if maxOfTwoInts(1, 1) != 1 {
		t.Errorf("Max should be 1")
	}
	if maxOfTwoInts(1, 0) != 1 {
		t.Errorf("Max should be 1")
	}
	if maxOfTwoInts(1, -1) != 1 {
		t.Errorf("Max should be 1")
	}
	if maxOfTwoInts(10000, -1) != 10000 {
		t.Errorf("Max should be 10000")
	}
}

func TestIsInArray(t *testing.T) {
	arr := []int{2, 3, 4, 5, 6, 7, 8, 9, 100, 200}
	for _, val := range arr {
		if !isInArray(arr, val) {
			t.Errorf("isInArray did not find %d", val)
		}
	}
	notarr := []int{-1, 0, 10, 150, 250, 10000}
	for _, val := range notarr {
		if isInArray(arr, val) {
			t.Errorf("isInArray failed on %d", val)
		}
	}
}

func TestIsInStringArray(t *testing.T) {
	arr := []string{"how", "now", "brown", "cow"}
	for _, val := range arr {
		if !isInStringArray(arr, val) {
			t.Errorf("isInStringArray did not find %s", val)
		}
	}
	notarr := []string{"these", "strings", "are", "not", "found", "COW", "Cow"}
	for _, val := range notarr {
		if isInStringArray(arr, val) {
			t.Errorf("isInStringArray failed on %s", val)
		}
	}
}

func TestCmpGroupOverlap(t *testing.T) {
	var o1Group, o2Group, o3Group Group
	var o1 GroupOverlapCount = GroupOverlapCount{cnt: 1}
	o1.group = &o1Group
	o1.group.ids = make([]int, 1)
	var o2 GroupOverlapCount = GroupOverlapCount{cnt: 2}
	o2.group = &o2Group
	o2.group.ids = make([]int, 2)
	var o3 GroupOverlapCount = GroupOverlapCount{cnt: 3}
	o3.group = &o3Group
	o3.group.ids = make([]int, 3)

	if cmpGroupOverlap(o1, o2) {
		t.Errorf("cmpGroupOverlap failed")
	}
	if cmpGroupOverlap(o1, o3) {
		t.Errorf("cmpGroupOverlap failed")
	}
	if cmpGroupOverlap(o2, o3) {
		t.Errorf("cmpGroupOverlap failed")
	}
}

func TestIsIdenticalGroup(t *testing.T) {
	var groupDb GroupDb

	nodeDb := initNodeDb()
	nodeDb.nodes = []Node{}
	groupDb.groups = []Group{}

	createConnectedNodes(&nodeDb, &groupDb, 5, "group1node")

	group1 := &groupDb.groups[0]
	var group2 Group = Group{groupName: group1.groupName, density: group1.density,
		totalNodeLinks: group1.totalNodeLinks, status: group1.status}
	group2.ids = make([]int, len(group1.ids))
	copy(group2.ids, group1.ids)
	group2.linkCounts = make([]int, len(group1.linkCounts))
	copy(group2.linkCounts, group1.linkCounts)

	if !isIdenticalGroup(group1, &group2) {
		t.Errorf("Groups are identical")
	}
	group2.ids = append(group2.ids, 999)
	if isIdenticalGroup(group1, &group2) {
		t.Errorf("Groups are not identical")
	}
}
